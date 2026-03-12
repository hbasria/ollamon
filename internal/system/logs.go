package system

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type RequestStat struct {
	Endpoint   string
	Count      int
	AvgLatency time.Duration
	LastStatus int
}

type LogTelemetry struct {
	RequestCount       int
	ErrorCount         int
	ErrorRate          float64
	AvgLatency         time.Duration
	LastEndpoint       string
	LastMethod         string
	LastStatus         int
	LastLatency        time.Duration
	LastSeen           time.Time
	TopEndpoints       []RequestStat
	SupportedTokenRate bool
	Window1m           WindowStat
	Window5m           WindowStat
}

type WindowStat struct {
	Count      int
	ErrorCount int
	ErrorRate  float64
	RPS        float64
}

func DefaultOllamaLogPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".ollama", "logs", "server.log")
}

func ReadRecentLogLines(path string, maxLines int) ([]string, string, error) {
	if path == "" {
		path = DefaultOllamaLogPath()
	}
	if path == "" {
		return nil, "", os.ErrNotExist
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, path, err
	}
	defer file.Close()

	lines := make([]string, 0, maxLines)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(lines) == maxLines {
			copy(lines, lines[1:])
			lines[len(lines)-1] = line
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, path, err
	}

	return lines, path, nil
}

func ReadLogTelemetry(path string, maxLines int) (LogTelemetry, []string, string, error) {
	lines, resolvedPath, err := ReadRecentLogLines(path, maxLines)
	if err != nil {
		return LogTelemetry{}, nil, resolvedPath, err
	}

	telemetry := ParseLogTelemetry(lines)
	return telemetry, lines, resolvedPath, nil
}

func ParseLogTelemetry(lines []string) LogTelemetry {
	type aggregate struct {
		count       int
		total       time.Duration
		lastStatus  int
		lastLatency time.Duration
	}

	var out LogTelemetry
	aggregates := map[string]aggregate{}
	var totalLatency time.Duration
	var latest time.Time

	for _, line := range lines {
		entry, ok := parseGINLine(line)
		if !ok {
			continue
		}
		if entry.Timestamp.After(latest) {
			latest = entry.Timestamp
		}
	}

	for _, line := range lines {
		entry, ok := parseGINLine(line)
		if !ok {
			continue
		}

		out.RequestCount++
		totalLatency += entry.Latency
		if entry.Status >= 400 {
			out.ErrorCount++
		}
		out.LastEndpoint = entry.Path
		out.LastMethod = entry.Method
		out.LastStatus = entry.Status
		out.LastLatency = entry.Latency
		out.LastSeen = entry.Timestamp

		agg := aggregates[entry.Path]
		agg.count++
		agg.total += entry.Latency
		agg.lastStatus = entry.Status
		agg.lastLatency = entry.Latency
		aggregates[entry.Path] = agg

		applyWindow(&out.Window1m, entry, latest, time.Minute)
		applyWindow(&out.Window5m, entry, latest, 5*time.Minute)
	}

	if out.RequestCount > 0 {
		out.AvgLatency = totalLatency / time.Duration(out.RequestCount)
		out.ErrorRate = float64(out.ErrorCount) * 100 / float64(out.RequestCount)
	}

	stats := make([]RequestStat, 0, len(aggregates))
	for endpoint, agg := range aggregates {
		stats = append(stats, RequestStat{
			Endpoint:   endpoint,
			Count:      agg.count,
			AvgLatency: agg.total / time.Duration(agg.count),
			LastStatus: agg.lastStatus,
		})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count == stats[j].Count {
			return stats[i].Endpoint < stats[j].Endpoint
		}
		return stats[i].Count > stats[j].Count
	})
	if len(stats) > 3 {
		stats = stats[:3]
	}
	out.TopEndpoints = stats

	return out
}

func applyWindow(window *WindowStat, entry ginEntry, latest time.Time, span time.Duration) {
	if latest.IsZero() || latest.Sub(entry.Timestamp) > span {
		return
	}
	window.Count++
	if entry.Status >= 400 {
		window.ErrorCount++
	}
	window.RPS = float64(window.Count) / span.Seconds()
	if window.Count > 0 {
		window.ErrorRate = float64(window.ErrorCount) * 100 / float64(window.Count)
	}
}

type ginEntry struct {
	Timestamp time.Time
	Status    int
	Latency   time.Duration
	Method    string
	Path      string
}

func parseGINLine(line string) (ginEntry, bool) {
	parts := strings.Split(line, "|")
	if len(parts) < 6 {
		return ginEntry{}, false
	}

	left := strings.TrimSpace(parts[0])
	timeText := strings.TrimSpace(strings.TrimPrefix(left, "[GIN]"))
	ts, err := time.Parse("2006/01/02 - 15:04:05", timeText)
	if err != nil {
		return ginEntry{}, false
	}

	status, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return ginEntry{}, false
	}

	latency, err := parseLatency(strings.TrimSpace(parts[2]))
	if err != nil {
		return ginEntry{}, false
	}

	method := strings.TrimSpace(parts[4])
	path := strings.TrimSpace(parts[5])
	path = strings.Trim(path, `"`)

	return ginEntry{
		Timestamp: ts,
		Status:    status,
		Latency:   latency,
		Method:    method,
		Path:      path,
	}, true
}

func parseLatency(raw string) (time.Duration, error) {
	raw = strings.ReplaceAll(strings.TrimSpace(raw), "µs", "us")
	if raw == "" {
		return 0, fmt.Errorf("empty latency")
	}
	return time.ParseDuration(raw)
}

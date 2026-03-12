package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/example/ollamon/internal/config"
	"github.com/example/ollamon/internal/ollama"
	"github.com/example/ollamon/internal/system"
	"github.com/example/ollamon/internal/tui"
	"github.com/example/ollamon/internal/util"
)

type sampleMsg struct {
	sample Sample
	err    error
}

type tickMsg time.Time

type Model struct {
	cfg    config.Config
	client *ollama.Client

	last        Sample
	err         error
	width       int
	height      int
	filter      string
	filterMode  bool
	cpuHistory  []float64
	memHistory  []float64
	diskHistory []float64
}

func New(cfg config.Config) Model {
	return Model{
		cfg:         cfg,
		client:      ollama.NewClient(cfg.BaseURL),
		width:       120,
		height:      40,
		cpuHistory:  make([]float64, 0, 40),
		memHistory:  make([]float64, 0, 40),
		diskHistory: make([]float64, 0, 40),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.refreshCmd(), tickCmd(m.cfg.Interval))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filterMode {
			switch msg.Type {
			case tea.KeyEsc:
				m.filterMode = false
				return m, nil
			case tea.KeyEnter:
				m.filterMode = false
				return m, nil
			case tea.KeyBackspace, tea.KeyDelete:
				if len(m.filter) > 0 {
					runes := []rune(m.filter)
					m.filter = string(runes[:len(runes)-1])
				}
				return m, nil
			default:
				if msg.Type == tea.KeyRunes {
					m.filter += msg.String()
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "/":
			m.filterMode = true
		case "esc":
			m.filter = ""
			m.filterMode = false
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case sampleMsg:
		m.last = msg.sample
		m.err = msg.err
		m.cpuHistory = pushHistory(m.cpuHistory, msg.sample.Host.CPUPercent, 48)
		m.memHistory = pushHistory(m.memHistory, msg.sample.Host.MemoryUsedPercent, 48)
		m.diskHistory = pushHistory(m.diskHistory, msg.sample.Host.DiskUsedPercent, 48)
	case tickMsg:
		return m, tea.Batch(m.refreshCmd(), tickCmd(m.cfg.Interval))
	}

	return m, nil
}

func (m Model) View() string {
	header := m.renderHeader()
	top := m.renderTopRow()
	middle := m.renderMiddleRow()
	bottom := m.renderBottomRow()
	help := tui.Help.Width(m.width).Render("q quit  / filter  esc clear filter")
	return strings.Join([]string{header, top, middle, bottom, help}, "\n")
}

func (m Model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		sample, err := CollectSample(m.cfg, m.client)
		return sampleMsg{sample: sample, err: err}
	}
}

func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func pushHistory(values []float64, next float64, limit int) []float64 {
	values = append(values, next)
	if len(values) > limit {
		values = values[len(values)-limit:]
	}
	return values
}

func (m Model) contentWidth() int {
	if m.width <= 4 {
		return 80
	}
	return m.width
}

func (m Model) renderHeader() string {
	status := tui.OK.Render("CONNECTED")
	if m.err != nil {
		status = tui.Err.Render("OFFLINE")
	} else if m.last.HealthError != "" {
		status = tui.Warn.Render("PARTIAL")
	}

	filter := tui.Subtle.Render("filter: all")
	if m.filter != "" || m.filterMode {
		value := m.filter
		if value == "" {
			value = "typing..."
		}
		filter = tui.Filter.Render("filter: " + value)
	}

	text := fmt.Sprintf("OLLAMON  %s  %s  %s", status, util.EmptyFallback(m.last.BaseURL, m.cfg.BaseURL), filter)
	return tui.Header.Width(m.contentWidth()).Render(text)
}

func (m Model) renderTopRow() string {
	left := m.renderOverview()
	right := m.renderTelemetry()
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) renderMiddleRow() string {
	left := m.renderRunning()
	right := m.renderInstalled()
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) renderBottomRow() string {
	left := m.renderInsightsPanel()
	right := m.renderLogPanel()
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) twoColumnWidths() (int, int) {
	total := m.contentWidth()
	if total < 100 {
		return total / 2, total - (total / 2)
	}
	left := total/2 - 1
	right := total - left
	return left, right
}

func (m Model) renderOverview() string {
	s := m.last
	leftW, _ := m.twoColumnWidths()

	cpuBar := util.PercentBar(s.Host.CPUPercent, 18, "█", "░")
	memBar := util.PercentBar(s.Host.MemoryUsedPercent, 18, "█", "░")
	diskBar := util.PercentBar(s.Host.DiskUsedPercent, 18, "█", "░")

	lines := []string{
		tui.Title.Render("Overview"),
		fmt.Sprintf("%s %s", tui.MetricLabel.Render("Host"), util.EmptyFallback(s.Host.Hostname, "-")),
		fmt.Sprintf("%s %s / %s", tui.MetricLabel.Render("Platform"), util.EmptyFallback(s.Host.Platform, "-"), util.EmptyFallback(s.Host.OS, "-")),
		fmt.Sprintf("%s %s", tui.MetricLabel.Render("Uptime"), util.HumanUptime(s.Host.Uptime)),
		fmt.Sprintf("%s %s", tui.MetricLabel.Render("GPU"), util.EmptyFallback(s.Host.GPUSummary, "-")),
		"",
		fmt.Sprintf("%s %5.1f%% %s", tui.MetricLabel.Render("CPU"), s.Host.CPUPercent, colorizePercent(s.Host.CPUPercent, cpuBar)),
		fmt.Sprintf("%s %5.1f%% %s", tui.MetricLabel.Render("MEM"), s.Host.MemoryUsedPercent, colorizePercent(s.Host.MemoryUsedPercent, memBar)),
		fmt.Sprintf("%s %5.1f%% %s", tui.MetricLabel.Render("DSK"), s.Host.DiskUsedPercent, colorizePercent(s.Host.DiskUsedPercent, diskBar)),
		"",
		fmt.Sprintf("%s %.1f / %.1f GB", tui.MetricLabel.Render("Memory"), s.Host.MemoryUsedGB, s.Host.MemoryTotalGB),
		fmt.Sprintf("%s %.1f / %.1f GB", tui.MetricLabel.Render("Disk"), s.Host.DiskUsedGB, s.Host.DiskTotalGB),
		fmt.Sprintf("%s %.2f / %.2f", tui.MetricLabel.Render("Load"), s.Host.Load1, s.Host.Load5),
		fmt.Sprintf("%s %d installed / %d running", tui.MetricLabel.Render("Models"), len(s.Installed), len(s.Running)),
	}

	if s.HealthError != "" {
		lines = append(lines, "", tui.Warn.Render("Ollama: "+s.HealthError))
	}

	return tui.Box.Width(leftW).Render(strings.Join(lines, "\n"))
}

func (m Model) renderTelemetry() string {
	_, rightW := m.twoColumnWidths()
	s := m.last

	cpuSpark := util.Sparkline(m.cpuHistory, 32)
	memSpark := util.Sparkline(m.memHistory, 32)
	diskSpark := util.Sparkline(m.diskHistory, 32)

	lines := []string{
		tui.Title.Render("Trends"),
		fmt.Sprintf("%s %s", tui.MetricLabel.Render("CPU"), tui.Accent.Render(cpuSpark)),
		fmt.Sprintf("%s %s", tui.MetricLabel.Render("MEM"), tui.Accent.Render(memSpark)),
		fmt.Sprintf("%s %s", tui.MetricLabel.Render("DSK"), tui.Accent.Render(diskSpark)),
		"",
		tui.Title.Render("Telemetry"),
		fmt.Sprintf("%s %s", tui.MetricLabel.Render("Token/s"), renderTokenRate(s)),
		fmt.Sprintf("%s %s", tui.MetricLabel.Render("Avg Lat"), renderLatency(s.LogStats.AvgLatency)),
		fmt.Sprintf("%s %s", tui.MetricLabel.Render("Last Req"), renderLastRequest(s)),
		fmt.Sprintf("%s %d", tui.MetricLabel.Render("Reqs"), s.LogStats.RequestCount),
		fmt.Sprintf("%s %.1f%%", tui.MetricLabel.Render("Errors"), s.LogStats.ErrorRate),
		fmt.Sprintf("%s %s", tui.MetricLabel.Render("1m Rate"), renderWindowStat(s.LogStats.Window1m)),
		fmt.Sprintf("%s %s", tui.MetricLabel.Render("5m Rate"), renderWindowStat(s.LogStats.Window5m)),
	}

	if len(s.LogStats.TopEndpoints) > 0 {
		lines = append(lines, "", tui.Section.Render("Top Endpoints"))
		for _, stat := range s.LogStats.TopEndpoints {
			lines = append(lines, fmt.Sprintf("%-12s %2d req  avg %s", stat.Endpoint, stat.Count, renderLatency(stat.AvgLatency)))
		}
	}

	if s.Time.IsZero() {
		lines = append(lines, "", tui.Dim.Render("İlk örnek bekleniyor..."))
	} else if s.LogStats.RequestCount == 0 {
		lines = append(lines, "", tui.Dim.Render("Log erişim kaydı bekleniyor."))
	}

	return tui.Box.Width(rightW).Render(strings.Join(lines, "\n"))
}

func (m Model) renderInsightsPanel() string {
	leftW, _ := m.twoColumnWidths()
	lines := []string{tui.Title.Render("Operational Insights"), m.renderInsights()}
	return tui.Box.Width(leftW).Render(strings.Join(lines, "\n"))
}

func (m Model) renderInsights() string {
	s := m.last
	insights := make([]string, 0, 6)

	if s.Host.MemoryUsedPercent > 85 {
		insights = append(insights, "• RAM kullanımı yüksek. Büyük model yüklemelerinde yavaşlama olabilir.")
	} else if s.Host.MemoryUsedPercent > 70 {
		insights = append(insights, "• RAM kullanımı orta-yüksek. Keep-alive süreleri gözden geçirilebilir.")
	} else {
		insights = append(insights, "• RAM tarafı şu an rahat görünüyor.")
	}

	if s.Host.DiskUsedPercent > 80 {
		insights = append(insights, "• Disk doluluğu kritik eşiğe yaklaşıyor. Eski modeller temizlenebilir.")
	}

	if len(s.Installed) > 0 {
		idx := 0
		for i := range s.Installed {
			if s.Installed[i].Size > s.Installed[idx].Size {
				idx = i
			}
		}
		insights = append(insights, fmt.Sprintf("• En büyük model: %s (%s)", s.Installed[idx].Name, util.BytesToHuman(s.Installed[idx].Size)))
	}

	if len(s.Running) > 0 {
		rm := s.Running[0]
		insights = append(insights, fmt.Sprintf("• Aktif örnek: %s, VRAM/RAM yükü yaklaşık %s", rm.Name, util.BytesToHuman(rm.SizeVRAM)))
	} else {
		insights = append(insights, "• Şu an bellekte aktif model görünmüyor.")
	}

	if s.LogStats.RequestCount > 0 {
		insights = append(insights, fmt.Sprintf("• Son %d log kaydında ortalama latency %s.", s.LogStats.RequestCount, renderLatency(s.LogStats.AvgLatency)))
		if len(s.LogStats.TopEndpoints) > 0 {
			insights = append(insights, fmt.Sprintf("• En yoğun endpoint: %s (%d istek).", s.LogStats.TopEndpoints[0].Endpoint, s.LogStats.TopEndpoints[0].Count))
		}
	}

	if m.filter != "" {
		insights = append(insights, fmt.Sprintf("• Filtre etkin: %q", m.filter))
	}

	return strings.Join(insights, "\n")
}

func (m Model) renderRunning() string {
	s := m.last
	leftW, _ := m.twoColumnWidths()
	lines := []string{tui.Title.Render("Running Models")}

	running := append([]ollama.RunningModel(nil), s.Running...)
	sort.Slice(running, func(i, j int) bool { return running[i].SizeVRAM > running[j].SizeVRAM })
	running = filterRunning(running, m.filter)

	if len(running) == 0 {
		lines = append(lines, tui.Dim.Render("Aktif model yok veya filtre eşleşmedi."))
		return tui.Box.Width(leftW).Render(strings.Join(lines, "\n"))
	}

	lines = append(lines, tui.Section.Render("NAME                        PARAMS      QUANT   VRAM       EXPIRES"))
	lines = append(lines, strings.Repeat("─", 74))

	for _, rm := range running {
		expires := "-"
		if !rm.ExpiresAt.IsZero() {
			expires = time.Until(rm.ExpiresAt).Round(time.Second).String()
		}
		lines = append(lines, fmt.Sprintf("%-27s %-11s %-7s %-10s %s",
			tui.Highlight.Render(util.Trim(rm.Name, 27)),
			util.Trim(util.EmptyFallback(rm.Details.ParameterSize, "-"), 11),
			util.Trim(util.EmptyFallback(rm.Details.QuantizationLevel, "-"), 7),
			util.Trim(util.BytesToHuman(rm.SizeVRAM), 10),
			expires,
		))
	}

	return tui.Box.Width(leftW).Render(strings.Join(lines, "\n"))
}

func (m Model) renderInstalled() string {
	s := m.last
	_, rightW := m.twoColumnWidths()
	lines := []string{tui.Title.Render("Installed Models")}

	models := append([]ollama.Model(nil), s.Installed...)
	sort.Slice(models, func(i, j int) bool { return models[i].Size > models[j].Size })
	models = filterInstalled(models, m.filter)

	if len(models) == 0 {
		lines = append(lines, tui.Dim.Render("Yüklü model bulunamadı veya filtre eşleşmedi."))
		return tui.Box.Width(rightW).Render(strings.Join(lines, "\n"))
	}

	lines = append(lines, tui.Section.Render("NAME                        SIZE       PARAMS      QUANT     MODIFIED"))
	lines = append(lines, strings.Repeat("─", 79))

	total := int64(0)
	for _, mod := range models {
		total += mod.Size
		lines = append(lines, fmt.Sprintf("%-27s %-10s %-11s %-9s %s",
			tui.Highlight.Render(util.Trim(mod.Name, 27)),
			util.Trim(util.BytesToHuman(mod.Size), 10),
			util.Trim(util.EmptyFallback(mod.Details.ParameterSize, "-"), 11),
			util.Trim(util.EmptyFallback(mod.Details.QuantizationLevel, "-"), 9),
			util.HumanAge(mod.ModifiedAt),
		))
	}

	lines = append(lines, "", fmt.Sprintf("Filtered cache size: %s", util.BytesToHuman(total)))
	return tui.Box.Width(rightW).Render(strings.Join(lines, "\n"))
}

func (m Model) renderLogPanel() string {
	_, rightW := m.twoColumnWidths()
	lines := []string{tui.Title.Render("Ollama Logs")}

	if m.last.LogPath != "" {
		lines = append(lines, tui.Subtle.Render(m.last.LogPath))
	}

	switch {
	case m.last.LogError != "":
		lines = append(lines, tui.Warn.Render(m.last.LogError))
	case len(m.last.LogLines) == 0:
		lines = append(lines, tui.Dim.Render("Log bulunamadı. OLLAMON_LOG_PATH ile özel yol verebilirsin."))
	default:
		logLines := m.last.LogLines
		if len(logLines) > 10 {
			logLines = logLines[len(logLines)-10:]
		}
		for _, line := range logLines {
			lines = append(lines, tui.LogLine.Render(util.Trim(line, max(24, rightW-6))))
		}
	}

	return tui.Box.Width(rightW).Render(strings.Join(lines, "\n"))
}

func filterInstalled(models []ollama.Model, filter string) []ollama.Model {
	if strings.TrimSpace(filter) == "" {
		return models
	}
	out := make([]ollama.Model, 0, len(models))
	needle := strings.ToLower(strings.TrimSpace(filter))
	for _, model := range models {
		haystack := strings.ToLower(strings.Join([]string{
			model.Name,
			model.Model,
			model.Details.ParameterSize,
			model.Details.QuantizationLevel,
			model.Details.Family,
		}, " "))
		if strings.Contains(haystack, needle) {
			out = append(out, model)
		}
	}
	return out
}

func filterRunning(models []ollama.RunningModel, filter string) []ollama.RunningModel {
	if strings.TrimSpace(filter) == "" {
		return models
	}
	out := make([]ollama.RunningModel, 0, len(models))
	needle := strings.ToLower(strings.TrimSpace(filter))
	for _, model := range models {
		haystack := strings.ToLower(strings.Join([]string{
			model.Name,
			model.Model,
			model.Details.ParameterSize,
			model.Details.QuantizationLevel,
			model.Details.Family,
		}, " "))
		if strings.Contains(haystack, needle) {
			out = append(out, model)
		}
	}
	return out
}

func colorizePercent(percent float64, bar string) string {
	switch {
	case percent >= 85:
		return tui.Err.Render(bar)
	case percent >= 70:
		return tui.Warn.Render(bar)
	default:
		return tui.OK.Render(bar)
	}
}

func renderTokenRate(s Sample) string {
	if s.LogStats.SupportedTokenRate {
		return tui.OK.Render("available")
	}
	return tui.Warn.Render("not in GIN access log")
}

func renderLatency(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	if d < time.Millisecond {
		return d.String()
	}
	return d.Round(time.Microsecond).String()
}

func renderLastRequest(s Sample) string {
	if s.LogStats.LastEndpoint == "" {
		return "-"
	}
	status := tui.OK.Render(fmt.Sprintf("%d", s.LogStats.LastStatus))
	if s.LogStats.LastStatus >= 400 {
		status = tui.Err.Render(fmt.Sprintf("%d", s.LogStats.LastStatus))
	}
	return fmt.Sprintf("%s %s %s", s.LogStats.LastMethod, s.LogStats.LastEndpoint, status)
}

func renderWindowStat(w system.WindowStat) string {
	if w.Count == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f rps  %d req  %.1f%% err", w.RPS, w.Count, w.ErrorRate)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BaseURL      string
	Interval     time.Duration
	RootDiskPath string
	Compact      bool
	LogPath      string
	Version      string
}

const defaultBaseURL = "http://127.0.0.1:11434"

func Load() Config {
	interval := 2 * time.Second
	if v := strings.TrimSpace(os.Getenv("OLLAMON_INTERVAL_MS")); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			interval = time.Duration(ms) * time.Millisecond
		}
	}

	compact := false
	if v := strings.TrimSpace(os.Getenv("OLLAMON_COMPACT")); v != "" {
		compact = v == "1" || strings.EqualFold(v, "true")
	}

	diskPath := strings.TrimSpace(os.Getenv("OLLAMON_DISK_PATH"))
	if diskPath == "" {
		diskPath = "/"
	}

	logPath := strings.TrimSpace(os.Getenv("OLLAMON_LOG_PATH"))

	return Config{
		BaseURL:      normalizeBaseURL(os.Getenv("OLLAMA_HOST")),
		Interval:     interval,
		RootDiskPath: diskPath,
		Compact:      compact,
		LogPath:      logPath,
	}
}

func normalizeBaseURL(s string) string {
	s = strings.TrimSpace(strings.TrimRight(s, "/"))
	if s == "" {
		return defaultBaseURL
	}
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return "http://" + s
	}
	return s
}

package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/example/ollamon/internal/config"
	"github.com/example/ollamon/internal/ollama"
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

	last   Sample
	err    error
	width  int
	height int
}

func New(cfg config.Config) Model {
	return Model{
		cfg:    cfg,
		client: ollama.NewClient(cfg.BaseURL),
		width:  120,
		height: 40,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.refreshCmd(), tickCmd(m.cfg.Interval))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case sampleMsg:
		m.last = msg.sample
		m.err = msg.err
	case tickMsg:
		return m, tea.Batch(m.refreshCmd(), tickCmd(m.cfg.Interval))
	}

	return m, nil
}

func (m Model) View() string {
	panels := []string{
		m.renderOverview(),
		m.renderRunning(),
		m.renderInstalled(),
	}
	return strings.Join(panels, "\n")
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

func (m Model) panelWidth() int {
	if m.width <= 4 {
		return 20
	}
	return m.width - 2
}

func (m Model) renderOverview() string {
	s := m.last

	status := tui.OK.Render("healthy")
	if m.err != nil {
		status = tui.Err.Render("degraded")
	} else if s.HealthError != "" {
		status = tui.Warn.Render("partial")
	}

	lines := []string{
		tui.Title.Render("Overview"),
		fmt.Sprintf("Target: %s", util.EmptyFallback(s.BaseURL, m.cfg.BaseURL)),
		fmt.Sprintf("Status: %s", status),
		fmt.Sprintf("Time: %s", util.SafeTime(s.Time)),
		fmt.Sprintf("Host: %s", util.EmptyFallback(s.Host.Hostname, "-")),
		fmt.Sprintf("Platform: %s / %s", util.EmptyFallback(s.Host.Platform, "-"), util.EmptyFallback(s.Host.OS, "-")),
		fmt.Sprintf("Uptime: %s", util.HumanUptime(s.Host.Uptime)),
		fmt.Sprintf("CPU: %.1f%%  Load: %.2f / %.2f", s.Host.CPUPercent, s.Host.Load1, s.Host.Load5),
		fmt.Sprintf("Memory: %.1f / %.1f GB (%.1f%%)", s.Host.MemoryUsedGB, s.Host.MemoryTotalGB, s.Host.MemoryUsedPercent),
		fmt.Sprintf("Disk: %.1f / %.1f GB (%.1f%%)", s.Host.DiskUsedGB, s.Host.DiskTotalGB, s.Host.DiskUsedPercent),
		fmt.Sprintf("GPU: %s", util.EmptyFallback(s.Host.GPUSummary, "-")),
		fmt.Sprintf("Installed: %d  Running: %d", len(s.Installed), len(s.Running)),
	}

	if s.HealthError != "" {
		lines = append(lines, tui.Warn.Render("Ollama: "+s.HealthError))
	}

	lines = append(lines, "", tui.Title.Render("Insights"), m.renderInsights())
	return tui.Box.Width(m.panelWidth()).Render(strings.Join(lines, "\n"))
}

func (m Model) renderInsights() string {
	s := m.last
	insights := make([]string, 0, 4)

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
	}

	insights = append(insights, "", tui.Dim.Render("RC sonrası aşamada token/s, request latency, filtreleme ve log paneli eklenebilir."))
	return strings.Join(insights, "\n")
}

func (m Model) renderRunning() string {
	s := m.last
	lines := []string{tui.Title.Render("Running models")}
	if len(s.Running) == 0 {
		lines = append(lines, tui.Dim.Render("Aktif model yok."))
		return tui.Box.Width(m.panelWidth()).Render(strings.Join(lines, "\n"))
	}

	lines = append(lines, "NAME                           PARAMS      QUANT   VRAM       EXPIRES")
	lines = append(lines, strings.Repeat("-", 78))

	running := append([]ollama.RunningModel(nil), s.Running...)
	sort.Slice(running, func(i, j int) bool { return running[i].SizeVRAM > running[j].SizeVRAM })

	for _, rm := range running {
		expires := "-"
		if !rm.ExpiresAt.IsZero() {
			expires = time.Until(rm.ExpiresAt).Round(time.Second).String()
		}
		lines = append(lines, fmt.Sprintf("%-30s %-11s %-7s %-10s %s",
			util.Trim(rm.Name, 30),
			util.Trim(util.EmptyFallback(rm.Details.ParameterSize, "-"), 11),
			util.Trim(util.EmptyFallback(rm.Details.QuantizationLevel, "-"), 7),
			util.Trim(util.BytesToHuman(rm.SizeVRAM), 10),
			expires,
		))
	}

	return tui.Box.Width(m.panelWidth()).Render(strings.Join(lines, "\n"))
}

func (m Model) renderInstalled() string {
	s := m.last
	lines := []string{tui.Title.Render("Installed models")}
	if len(s.Installed) == 0 {
		lines = append(lines, tui.Dim.Render("Yüklü model bulunamadı."))
		return tui.Box.Width(m.panelWidth()).Render(strings.Join(lines, "\n"))
	}

	models := append([]ollama.Model(nil), s.Installed...)
	sort.Slice(models, func(i, j int) bool { return models[i].Size > models[j].Size })

	lines = append(lines, "NAME                           SIZE       PARAMS      QUANT     MODIFIED")
	lines = append(lines, strings.Repeat("-", 92))

	total := int64(0)
	for _, mod := range models {
		total += mod.Size
		lines = append(lines, fmt.Sprintf("%-30s %-10s %-11s %-9s %s",
			util.Trim(mod.Name, 30),
			util.Trim(util.BytesToHuman(mod.Size), 10),
			util.Trim(util.EmptyFallback(mod.Details.ParameterSize, "-"), 11),
			util.Trim(util.EmptyFallback(mod.Details.QuantizationLevel, "-"), 9),
			util.HumanAge(mod.ModifiedAt),
		))
	}

	lines = append(lines, "", fmt.Sprintf("Toplam model cache boyutu: %s", util.BytesToHuman(total)))
	return tui.Box.Width(m.panelWidth()).Render(strings.Join(lines, "\n"))
}

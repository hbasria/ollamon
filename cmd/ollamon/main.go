package main

import (
	"fmt"
	"os"

	"github.com/example/ollamon/internal/app"
	"github.com/example/ollamon/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func main() {
	cfg := config.Load()
	cfg.Version = version

	m := app.New(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ollamon error: %v\n", err)
		os.Exit(1)
	}
}

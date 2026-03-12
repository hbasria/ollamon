package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/ollamon/internal/config"
	"github.com/example/ollamon/internal/ollama"
	"github.com/example/ollamon/internal/system"
)

type Sample struct {
	Time        time.Time
	BaseURL     string
	Healthy     bool
	HealthError string
	Host        system.Stats
	Installed   []ollama.Model
	Running     []ollama.RunningModel
	LogPath     string
	LogLines    []string
	LogError    string
	LogStats    system.LogTelemetry
}

func CollectSample(cfg config.Config, client *ollama.Client) (Sample, error) {
	s := Sample{Time: time.Now(), BaseURL: client.BaseURL}

	hs, _ := system.Collect(cfg.RootDiskPath)
	s.Host = hs

	logStats, logLines, logPath, logErr := system.ReadLogTelemetry(cfg.LogPath, 10)
	s.LogLines = logLines
	s.LogPath = logPath
	s.LogStats = logStats
	if logErr != nil {
		s.LogError = logErr.Error()
	}

	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	var (
		installed []ollama.Model
		running   []ollama.RunningModel
		errTags   error
		errPS     error
		wg        sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		installed, errTags = client.Tags(ctx)
	}()

	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		running, errPS = client.PS(ctx)
	}()

	wg.Wait()

	s.Installed = installed
	s.Running = running
	s.Healthy = errTags == nil || errPS == nil

	switch {
	case errTags == nil && errPS == nil:
		return s, nil
	case errTags != nil && errPS != nil:
		s.HealthError = fmt.Sprintf("tags: %v | ps: %v", errTags, errPS)
		return s, fmt.Errorf("ollama unavailable: %s", s.HealthError)
	case errTags != nil:
		s.HealthError = errTags.Error()
		return s, nil
	default:
		s.HealthError = errPS.Error()
		return s, nil
	}
}

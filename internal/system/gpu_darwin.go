//go:build darwin

package system

import (
	"encoding/json"
	"os/exec"
	"strings"
)

type agputopPayload struct {
	GPU struct {
		FrequencyMHz   float64 `json:"frequency_mhz"`
		PowerW         float64 `json:"power_w"`
		UtilizationPct float64 `json:"utilization_pct"`
	} `json:"gpu"`
	System struct {
		Chip     string  `json:"chip"`
		MemoryGB float64 `json:"memory_gb"`
	} `json:"system"`
}

func ProbeGPUStats() GPUStats {
	if out, err := exec.Command("agputop", "--json").Output(); err == nil {
		var payload agputopPayload
		if err := json.Unmarshal(out, &payload); err == nil {
			return GPUStats{
				Name:           strings.TrimSpace(payload.System.Chip),
				UtilizationPct: payload.GPU.UtilizationPct,
				PowerW:         payload.GPU.PowerW,
				FrequencyMHz:   payload.GPU.FrequencyMHz,
				MemoryTotalGB:  payload.System.MemoryGB,
				Available:      true,
				Backend:        "agputop",
			}
		}
	}

	out, err := exec.Command("/usr/sbin/system_profiler", "SPDisplaysDataType").CombinedOutput()
	if err != nil {
		return GPUStats{}
	}
	text := string(out)
	switch {
	case strings.Contains(text, "Apple M3"):
		return GPUStats{Name: "Apple M3 family", Available: true, Backend: "system_profiler"}
	case strings.Contains(text, "Apple M2"):
		return GPUStats{Name: "Apple M2 family", Available: true, Backend: "system_profiler"}
	case strings.Contains(text, "Apple M1"):
		return GPUStats{Name: "Apple M1 family", Available: true, Backend: "system_profiler"}
	default:
		return GPUStats{Name: "Apple GPU detected", Available: true, Backend: "system_profiler"}
	}
}

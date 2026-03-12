//go:build linux

package system

import (
	"os/exec"
	"strconv"
	"strings"
)

func ProbeGPUStats() GPUStats {
	if out, err := exec.Command("bash", "-lc", "command -v nvidia-smi >/dev/null 2>&1 && nvidia-smi --query-gpu=name,utilization.gpu,memory.used,memory.total --format=csv,noheader,nounits | head -n 1").CombinedOutput(); err == nil {
		line := strings.TrimSpace(string(out))
		if line != "" {
			parts := strings.Split(line, ",")
			if len(parts) >= 4 {
				util, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				memUsedMB, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
				memTotalMB, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
				return GPUStats{
					Name:           strings.TrimSpace(parts[0]),
					UtilizationPct: util,
					MemoryUsedGB:   memUsedMB / 1024,
					MemoryTotalGB:  memTotalMB / 1024,
					Available:      true,
					Backend:        "nvidia-smi",
				}
			}
		}
	}
	return GPUStats{}
}

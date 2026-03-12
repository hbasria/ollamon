//go:build linux

package system

import (
	"os/exec"
	"strings"
)

func ProbeGPUSummary() string {
	if out, err := exec.Command("bash", "-lc", "command -v nvidia-smi >/dev/null 2>&1 && nvidia-smi --query-gpu=name,utilization.gpu,memory.used,memory.total --format=csv,noheader,nounits | head -n 1").CombinedOutput(); err == nil {
		line := strings.TrimSpace(string(out))
		if line != "" {
			return "linux gpu: " + line
		}
	}
	return "linux gpu: unavailable"
}

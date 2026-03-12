//go:build darwin

package system

import (
	"os/exec"
	"strings"
)

func ProbeGPUSummary() string {
	out, err := exec.Command("/usr/sbin/system_profiler", "SPDisplaysDataType").CombinedOutput()
	if err != nil {
		return "mac gpu: unavailable"
	}
	text := string(out)
	switch {
	case strings.Contains(text, "Apple M3"):
		return "mac gpu: Apple M3 family detected"
	case strings.Contains(text, "Apple M2"):
		return "mac gpu: Apple M2 family detected"
	case strings.Contains(text, "Apple M1"):
		return "mac gpu: Apple M1 family detected"
	default:
		return "mac gpu: detected"
	}
}

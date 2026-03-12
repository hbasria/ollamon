package util

import (
	"strings"
)

var sparkRunes = []rune("▁▂▃▄▅▆▇█")

func Sparkline(values []float64, width int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}
	if len(values) > width {
		values = values[len(values)-width:]
	}

	max := 0.0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if max <= 0 {
		return strings.Repeat(string(sparkRunes[0]), len(values))
	}

	out := make([]rune, 0, len(values))
	last := float64(len(sparkRunes) - 1)
	for _, v := range values {
		idx := int((v / max) * last)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkRunes) {
			idx = len(sparkRunes) - 1
		}
		out = append(out, sparkRunes[idx])
	}
	return string(out)
}

func PercentBar(percent float64, width int, filled string, empty string) string {
	if width <= 0 {
		return ""
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filledCount := int((percent / 100) * float64(width))
	if filledCount > width {
		filledCount = width
	}
	return strings.Repeat(filled, filledCount) + strings.Repeat(empty, width-filledCount)
}

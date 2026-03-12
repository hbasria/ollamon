package system

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

type Stats struct {
	Hostname          string
	OS                string
	Platform          string
	Uptime            uint64
	CPUPercent        float64
	Load1             float64
	Load5             float64
	MemoryUsedGB      float64
	MemoryTotalGB     float64
	MemoryUsedPercent float64
	DiskUsedGB        float64
	DiskTotalGB       float64
	DiskUsedPercent   float64
	GPUSummary        string
}

func Collect(rootDiskPath string) (Stats, error) {
	var s Stats

	info, _ := host.Info()
	s.Hostname = info.Hostname
	s.OS = info.OS
	s.Platform = info.Platform
	s.Uptime = info.Uptime

	if values, err := cpu.Percent(250*time.Millisecond, false); err == nil && len(values) > 0 {
		s.CPUPercent = values[0]
	}
	if avg, err := load.Avg(); err == nil {
		s.Load1 = avg.Load1
		s.Load5 = avg.Load5
	}
	if vm, err := mem.VirtualMemory(); err == nil {
		s.MemoryUsedGB = toGB(vm.Used)
		s.MemoryTotalGB = toGB(vm.Total)
		s.MemoryUsedPercent = vm.UsedPercent
	}
	if du, err := disk.Usage(rootDiskPath); err == nil {
		s.DiskUsedGB = toGB(du.Used)
		s.DiskTotalGB = toGB(du.Total)
		s.DiskUsedPercent = du.UsedPercent
	}

	s.GPUSummary = ProbeGPUSummary()
	return s, nil
}

func toGB(v uint64) float64 {
	return float64(v) / 1024 / 1024 / 1024
}

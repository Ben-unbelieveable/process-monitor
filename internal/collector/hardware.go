package collector

import (
	"os"
	"runtime"

	"github.com/liubo/process-monitor/internal/models"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// GetSystemHardware 采集当前系统整体硬件配置。
func GetSystemHardware(gpuType string) models.SystemHardware {
	hw := models.SystemHardware{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
	if hostname, err := os.Hostname(); err == nil {
		hw.Hostname = hostname
	}
	if infos, err := cpu.Info(); err == nil && len(infos) > 0 {
		hw.CPUModel = infos[0].ModelName
	}
	if cores, err := cpu.Counts(false); err == nil {
		hw.CPUCores = cores
	}
	if threads, err := cpu.Counts(true); err == nil {
		hw.CPUThreads = threads
	}
	if vm, err := mem.VirtualMemory(); err == nil {
		hw.MemTotalGB = float64(vm.Total) / (1024 * 1024 * 1024)
	}

	switch gpuType {
	case "nvidia":
		hw.GPUs = nvidiaGPUDevices()
	case "apple":
		hw.GPUs = appleGPUDevices()
	case "both":
		if gpus := nvidiaGPUDevices(); len(gpus) > 0 {
			hw.GPUs = gpus
		} else {
			hw.GPUs = appleGPUDevices()
		}
	}
	return hw
}

func nvidiaGPUDevices() []models.GPUInfo {
	nvidia := GetGPUInfoNvidia()
	devices := make([]models.GPUInfo, 0, len(nvidia))
	for _, g := range nvidia {
		devices = append(devices, models.GPUInfo{
			Index:              g.GPUIndex,
			Name:               g.GPUName,
			MemoryTotalMB:      g.MemoryTotalMB,
			MemoryUsedMB:       g.GPUMemoryMB,
			UtilizationPercent: g.GPUUtilization,
			DriverVersion:      g.DriverVersion,
		})
	}
	return devices
}

func appleGPUDevices() []models.GPUInfo {
	apple := GetGPUInfoApple()
	devices := make([]models.GPUInfo, 0, len(apple))
	for i, g := range apple {
		devices = append(devices, models.GPUInfo{
			Index:              i,
			Name:               g.GPUName,
			MemoryTotalMB:      g.MemoryTotalMB,
			MemoryUsedMB:       g.GPUMemoryMB,
			UtilizationPercent: g.GPUUtilization,
		})
	}
	return devices
}

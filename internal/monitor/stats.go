package monitor

import (
	"github.com/liubo/process-monitor/internal/collector"
	"github.com/liubo/process-monitor/internal/config"
	"github.com/liubo/process-monitor/internal/models"
)

// gpuSnapshot 保存当前轮次各 GPU 全局状态。
type gpuSnapshot struct {
	nvidia map[int]collector.NvidiaGPUInfo
	apple  map[int]collector.AppleGPUInfo
}

func resolveGPUSnapshot(cfg *config.Config) gpuSnapshot {
	snap := gpuSnapshot{}
	if cfg.GPUType == "nvidia" || cfg.GPUType == "both" {
		snap.nvidia = collector.GetGPUMapNvidia()
	}
	if cfg.GPUType == "apple" || cfg.GPUType == "both" {
		snap.apple = collector.GetGPUMapApple()
	}
	return snap
}

// buildProcessRows 根据进程 GPU 占用展开为多行（每块 GPU 一行）。
func buildProcessRows(
	timestamp string,
	pid int32,
	name string,
	cpu, mem float64,
	procGPUs map[int]float64,
	snap gpuSnapshot,
	cfg *config.Config,
) []models.ProcessStats {
	if len(procGPUs) == 0 {
		return []models.ProcessStats{{
			Timestamp:      timestamp,
			PID:            pid,
			ProcessName:    name,
			CPUPercent:     cpu,
			MemoryGB:       mem,
			GPUIndex:       -1,
			GPUMemoryMB:    0,
			GPUUtilization: 0,
			GPUName:        "N/A",
		}}
	}

	var rows []models.ProcessStats
	for idx, memMB := range procGPUs {
		if !cfg.GPUIndexAllowed(idx) {
			continue
		}
		row := models.ProcessStats{
			Timestamp:   timestamp,
			PID:         pid,
			ProcessName: name,
			CPUPercent:  cpu,
			MemoryGB:    mem,
			GPUIndex:    idx,
			GPUMemoryMB: memMB,
		}
		if g, ok := snap.nvidia[idx]; ok {
			row.GPUUtilization = g.GPUUtilization
			row.GPUName = g.GPUName
		} else if g, ok := snap.apple[idx]; ok {
			row.GPUUtilization = g.GPUUtilization
			row.GPUName = g.GPUName
		} else {
			row.GPUName = "N/A"
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return []models.ProcessStats{{
			Timestamp:      timestamp,
			PID:            pid,
			ProcessName:    name,
			CPUPercent:     cpu,
			MemoryGB:       mem,
			GPUIndex:       -1,
			GPUMemoryMB:    0,
			GPUUtilization: 0,
			GPUName:        "N/A",
		}}
	}
	return rows
}

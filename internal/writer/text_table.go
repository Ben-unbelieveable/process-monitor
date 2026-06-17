package writer

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/liubo/process-monitor/internal/models"
)

const (
	colTime    = 19
	colPID     = 8
	colName    = 18
	colCPU     = 8
	colMem     = 10
	colGPUIdx  = 5
	colGPUMem  = 10
	colGPUUtil = 8
	colGPUName = 28
)

func textTableHeader() string {
	return fmt.Sprintf("%-*s %*s %-*s %*s %*s %*s %*s %*s %-*s",
		colTime, "timestamp",
		colPID, "pid",
		colName, "process_name",
		colCPU, "cpu_%",
		colMem, "memory_gb",
		colGPUIdx, "gpu",
		colGPUMem, "gpu_mem_mb",
		colGPUUtil, "gpu_%",
		colGPUName, "gpu_name",
	)
}

func textTableSeparator() string {
	widths := []int{colTime, colPID, colName, colCPU, colMem, colGPUIdx, colGPUMem, colGPUUtil, colGPUName}
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("-", w)
	}
	return strings.Join(parts, " ")
}

func formatTextRow(s models.ProcessStats) string {
	gpuIdx := "-"
	if s.GPUIndex >= 0 {
		gpuIdx = fmt.Sprintf("%d", s.GPUIndex)
	}
	return fmt.Sprintf("%-*s %*d %-*s %*s %*s %*s %*s %*s %-*s",
		colTime, s.Timestamp,
		colPID, s.PID,
		colName, clipRunes(s.ProcessName, colName),
		colCPU, fmt.Sprintf("%.1f", s.CPUPercent),
		colMem, fmt.Sprintf("%.3f", s.MemoryGB),
		colGPUIdx, gpuIdx,
		colGPUMem, fmt.Sprintf("%.1f", s.GPUMemoryMB),
		colGPUUtil, fmt.Sprintf("%.1f", s.GPUUtilization),
		colGPUName, clipRunes(s.GPUName, colGPUName),
	)
}

func clipRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	runes := []rune(s)
	return string(runes[:max-1]) + "…"
}

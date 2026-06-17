// Package collector 提供 CPU/内存及 GPU 资源采集能力。
package collector

import (
	"strings"
	"time"

	"github.com/liubo/process-monitor/internal/models"
	"github.com/shirou/gopsutil/v4/process"
)

// ProcessStatResult 表示单个进程的 CPU/内存采集结果。
type ProcessStatResult struct {
	PID         int32
	ProcessName string
	CPUPercent  float64
	MemoryGB    float64
}

// GetProcessStats 获取单个进程的 CPU 和内存使用率。
func GetProcessStats(pid int32, name string) *ProcessStatResult {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return nil
	}

	memInfo, err := proc.MemoryInfo()
	if err != nil {
		return nil
	}

	cpuPercent, err := proc.Percent(time.Millisecond * 100)
	if err != nil {
		cpuPercent = 0
	}

	return &ProcessStatResult{
		PID:         pid,
		ProcessName: name,
		CPUPercent:  cpuPercent,
		MemoryGB:    float64(memInfo.RSS) / (1024 * 1024 * 1024),
	}
}

// GetProcessCmdline 获取进程完整命令行。
func GetProcessCmdline(pid int32) string {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return ""
	}
	parts, err := proc.CmdlineSlice()
	if err != nil || len(parts) == 0 {
		line, err := proc.Cmdline()
		if err != nil {
			return ""
		}
		return line
	}
	return strings.Join(parts, " ")
}

// GetAllProcesses 枚举系统所有进程，用于自动模式下的内存过滤。
func GetAllProcesses() []models.ProcInfo {
	procs, err := process.Processes()
	if err != nil {
		return nil
	}

	var results []models.ProcInfo
	for _, proc := range procs {
		pid := proc.Pid
		name, err := proc.Name()
		if err != nil {
			continue
		}
		memInfo, err := proc.MemoryInfo()
		if err != nil {
			continue
		}
		cmdline := ""
		if parts, err := proc.CmdlineSlice(); err == nil && len(parts) > 0 {
			cmdline = strings.Join(parts, " ")
		} else if line, err := proc.Cmdline(); err == nil {
			cmdline = line
		}
		results = append(results, models.ProcInfo{
			PID:      pid,
			Name:     name,
			Cmdline:  cmdline,
			MemoryGB: float64(memInfo.RSS) / (1024 * 1024 * 1024),
		})
	}
	return results
}

// FilterByMemory 按内存阈值过滤进程列表。
func FilterByMemory(procs []models.ProcInfo, thresholdGB float64) []models.ProcInfo {
	var matched []models.ProcInfo
	for _, p := range procs {
		if p.MemoryGB >= thresholdGB {
			matched = append(matched, p)
		}
	}
	return matched
}

// FilterByPIDs 按 PID 列表过滤。
func FilterByPIDs(procs []models.ProcInfo, targetPIDs []int32) []models.ProcInfo {
	pidSet := make(map[int32]struct{}, len(targetPIDs))
	for _, pid := range targetPIDs {
		pidSet[pid] = struct{}{}
	}
	var matched []models.ProcInfo
	for _, p := range procs {
		if _, ok := pidSet[p.PID]; ok {
			matched = append(matched, p)
		}
	}
	return matched
}

// FilterByNames 按进程名模糊匹配（不区分大小写）。
func FilterByNames(procs []models.ProcInfo, names []string) []models.ProcInfo {
	var matched []models.ProcInfo
	for _, p := range procs {
		for _, name := range names {
			if strings.Contains(strings.ToLower(p.Name), strings.ToLower(name)) {
				matched = append(matched, p)
				break
			}
		}
	}
	return matched
}

// FilterByCmdline 按命令行关键词模糊匹配（不区分大小写）。
func FilterByCmdline(procs []models.ProcInfo, keywords []string) []models.ProcInfo {
	var matched []models.ProcInfo
	for _, p := range procs {
		cmdLower := strings.ToLower(p.Cmdline)
		if cmdLower == "" {
			continue
		}
		for _, kw := range keywords {
			if strings.Contains(cmdLower, strings.ToLower(kw)) {
				matched = append(matched, p)
				break
			}
		}
	}
	return matched
}

// LookupProcByPID 按 PID 查找进程信息，不存在时返回最小信息。
func LookupProcByPID(all []models.ProcInfo, pid int32) models.ProcInfo {
	for _, p := range all {
		if p.PID == pid {
			return p
		}
	}
	return models.ProcInfo{PID: pid, Name: "unknown"}
}

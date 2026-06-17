package collector

import (
	"fmt"

	"github.com/liubo/process-monitor/internal/models"
	"github.com/shirou/gopsutil/v4/process"
)

// BuildParentMap 构建 pid -> ppid 映射。
func BuildParentMap() map[int32]int32 {
	procs, err := process.Processes()
	if err != nil {
		return nil
	}
	parentOf := make(map[int32]int32, len(procs))
	for _, proc := range procs {
		ppid, err := proc.Ppid()
		if err != nil {
			continue
		}
		parentOf[proc.Pid] = ppid
	}
	return parentOf
}

// BuildChildMap 构建 ppid -> 子 pid 列表映射。
func BuildChildMap() map[int32][]int32 {
	parentOf := BuildParentMap()
	childrenOf := make(map[int32][]int32)
	for pid, ppid := range parentOf {
		childrenOf[ppid] = append(childrenOf[ppid], pid)
	}
	return childrenOf
}

// TreePIDs 返回 root 及其所有子孙进程 PID（含 root）。
func TreePIDs(root int32, childrenOf map[int32][]int32) []int32 {
	if root <= 0 {
		return nil
	}
	seen := map[int32]struct{}{root: {}}
	queue := []int32{root}
	result := []int32{root}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range childrenOf[cur] {
			if _, ok := seen[child]; ok {
				continue
			}
			seen[child] = struct{}{}
			result = append(result, child)
			queue = append(queue, child)
		}
	}
	return result
}

// FilterTopLevel 去掉被其他匹配项包含的子孙进程，避免进程树重复统计。
func FilterTopLevel(matches []models.ProcInfo, parentOf map[int32]int32) []models.ProcInfo {
	if len(matches) <= 1 {
		return matches
	}
	matchSet := make(map[int32]struct{}, len(matches))
	for _, m := range matches {
		matchSet[m.PID] = struct{}{}
	}

	var tops []models.ProcInfo
	for _, m := range matches {
		if hasMatchedAncestor(m.PID, matchSet, parentOf) {
			continue
		}
		tops = append(tops, m)
	}
	return tops
}

func hasMatchedAncestor(pid int32, matchSet map[int32]struct{}, parentOf map[int32]int32) bool {
	for {
		ppid, ok := parentOf[pid]
		if !ok || ppid <= 0 {
			return false
		}
		if _, matched := matchSet[ppid]; matched {
			return true
		}
		pid = ppid
	}
}

// AggregateStats 汇总多个 PID 的 CPU 和内存。
func AggregateStats(pids []int32, rootName string) (cpu, memGB float64, name string) {
	name = rootName
	for _, pid := range pids {
		procName := rootName
		if proc, err := process.NewProcess(pid); err == nil {
			if n, err := proc.Name(); err == nil {
				procName = n
			}
		}
		if stats := GetProcessStats(pid, procName); stats != nil {
			cpu += stats.CPUPercent
			memGB += stats.MemoryGB
		}
	}
	if name == "" {
		name = "unknown"
	}
	if len(pids) > 1 {
		name = fmt.Sprintf("%s (tree:%d)", rootName, len(pids)-1)
	}
	return cpu, memGB, name
}

// AggregateTreeGPUs 汇总进程树在各 NVIDIA GPU 上的显存（MB）。
func AggregateTreeGPUs(pids []int32) map[int]float64 {
	result := make(map[int]float64)
	for _, pid := range pids {
		for _, g := range ProcessGPUMemoryAll(pid) {
			result[g.GPUIndex] += g.GPUMemoryMB
		}
	}
	return result
}

// CollectProcGPUs 采集单个进程在各 NVIDIA GPU 上的显存占用。
func CollectProcGPUs(pid int32, gpuType string) map[int]float64 {
	result := make(map[int]float64)
	if gpuType == "nvidia" || gpuType == "both" {
		for _, g := range ProcessGPUMemoryAll(pid) {
			result[g.GPUIndex] += g.GPUMemoryMB
		}
	}
	return result
}

// CollectTreeGPUs 采集进程树在各 NVIDIA GPU 上的显存占用。
func CollectTreeGPUs(pids []int32, gpuType string) map[int]float64 {
	if gpuType == "nvidia" || gpuType == "both" {
		return AggregateTreeGPUs(pids)
	}
	return nil
}

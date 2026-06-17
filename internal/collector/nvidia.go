package collector

import (
	"os/exec"
	"strconv"
	"strings"
)

// NvidiaGPUInfo 表示单块 NVIDIA GPU 的信息。
type NvidiaGPUInfo struct {
	GPUName         string
	GPUMemoryMB     float64
	MemoryTotalMB   float64
	GPUUtilization  float64
	GPUIndex        int
	DriverVersion   string
}

// IsNvidiaAvailable 检测系统是否可用 nvidia-smi 命令。
func IsNvidiaAvailable() bool {
	_, err := exec.LookPath("nvidia-smi")
	return err == nil
}

// GetGPUInfoNvidia 通过 nvidia-smi 获取所有 GPU 信息。
func GetGPUInfoNvidia() []NvidiaGPUInfo {
	if !IsNvidiaAvailable() {
		return nil
	}

	out, err := exec.Command("nvidia-smi",
		"--query-gpu=index,name,memory.total,memory.used,utilization.gpu,driver_version",
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		return nil
	}

	var results []NvidiaGPUInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.Split(line, ", ")
		if len(parts) < 6 {
			continue
		}
		idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		memTotal, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		memUsed, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		util, _ := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
		results = append(results, NvidiaGPUInfo{
			GPUIndex:       idx,
			GPUName:        strings.TrimSpace(parts[1]),
			MemoryTotalMB:  memTotal,
			GPUMemoryMB:    memUsed,
			GPUUtilization: util,
			DriverVersion:  strings.TrimSpace(parts[5]),
		})
	}
	return results
}

// ProcessGPUMemoryAll 获取指定进程在各 GPU 上的显存使用（MB）。
func ProcessGPUMemoryAll(pid int32) []NvidiaGPUInfo {
	if !IsNvidiaAvailable() {
		return nil
	}

	out, err := exec.Command("nvidia-smi",
		"--query-compute-apps=gpu_index,pid,used_gpu_memory",
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		return nil
	}

	pidStr := strconv.Itoa(int(pid))
	var results []NvidiaGPUInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.Split(line, ", ")
		if len(parts) < 3 {
			continue
		}
		if strings.TrimSpace(parts[1]) != pidStr {
			continue
		}
		idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		memMB, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		results = append(results, NvidiaGPUInfo{
			GPUIndex:    idx,
			GPUMemoryMB: memMB,
		})
	}
	return results
}

// ProcessGPUMemory 获取指定进程在单块 GPU 上的显存（兼容旧调用，返回合计第一条）。
func ProcessGPUMemory(pid int32) *NvidiaGPUInfo {
	all := ProcessGPUMemoryAll(pid)
	if len(all) == 0 {
		return nil
	}
	total := 0.0
	for _, g := range all {
		total += g.GPUMemoryMB
	}
	return &NvidiaGPUInfo{GPUMemoryMB: total}
}

// GetGPUMapNvidia 返回 index -> GPU 信息映射。
func GetGPUMapNvidia() map[int]NvidiaGPUInfo {
	gpus := GetGPUInfoNvidia()
	m := make(map[int]NvidiaGPUInfo, len(gpus))
	for _, g := range gpus {
		m[g.GPUIndex] = g
	}
	return m
}

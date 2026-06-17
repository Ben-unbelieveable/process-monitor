// Package models 定义进程监控的数据结构。
package models

// ProcessStats 表示单个进程在一轮采集中的资源使用快照。
// 多 GPU 场景下，同一进程可能对应多行（每块 GPU 一行）。
type ProcessStats struct {
	Timestamp      string  `json:"timestamp"`
	PID            int32   `json:"pid"`
	ProcessName    string  `json:"process_name"`
	CPUPercent     float64 `json:"cpu_percent"`
	MemoryGB       float64 `json:"memory_gb"`
	GPUIndex       int     `json:"gpu_index"` // -1 表示未使用 GPU
	GPUMemoryMB    float64 `json:"gpu_memory_mb"`
	GPUUtilization float64 `json:"gpu_utilization"`
	GPUName        string  `json:"gpu_name"`
}

// Header 返回 CSV/TSV 列名。
func Header() []string {
	return []string{
		"timestamp", "pid", "process_name", "cpu_percent",
		"memory_gb", "gpu_index", "gpu_memory_mb", "gpu_utilization", "gpu_name",
	}
}

// GPUInfo 表示单块 GPU 的全局状态。
type GPUInfo struct {
	Index              int
	Name               string
	MemoryTotalMB      float64
	MemoryUsedMB       float64
	UtilizationPercent float64
	DriverVersion      string
}

// EmptyGPU 返回表示无 GPU 信息的默认值。
func EmptyGPU() GPUInfo {
	return GPUInfo{Index: -1, Name: "N/A"}
}

// ProcInfo 表示进程基本信息，用于目标进程筛选。
type ProcInfo struct {
	PID      int32
	Name     string
	Cmdline  string
	MemoryGB float64
}

// SystemHardware 表示系统整体硬件配置。
type SystemHardware struct {
	Hostname   string
	OS         string
	Arch       string
	CPUModel   string
	CPUCores   int
	CPUThreads int
	MemTotalGB float64
	GPUs       []GPUInfo
}

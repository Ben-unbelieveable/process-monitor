package models

// ContainerInfo 表示容器运行环境及 cgroup 资源限额。
type ContainerInfo struct {
	InContainer   bool
	Runtime       string  // docker / kubernetes / podman / containerd / none
	CgroupPath    string
	PodName       string
	PodNamespace  string
	NodeName      string
	MemLimitBytes uint64  // 0 表示无限制
	MemUsageBytes uint64
	MemUsagePct   float64 // 相对限额百分比；-1 表示无限额
	CPUCoresLimit float64 // 0 表示无限制
}

// MemLimitGB 返回内存限额（GB），无限额时返回 0。
func (c ContainerInfo) MemLimitGB() float64 {
	if c.MemLimitBytes == 0 {
		return 0
	}
	return float64(c.MemLimitBytes) / (1024 * 1024 * 1024)
}

// MemUsageGB 返回当前内存用量（GB）。
func (c ContainerInfo) MemUsageGB() float64 {
	return float64(c.MemUsageBytes) / (1024 * 1024 * 1024)
}

package monitor

import (
	"fmt"
	"os"

	"github.com/liubo/process-monitor/internal/config"
	"github.com/liubo/process-monitor/internal/models"
)

// Alerter 在指标超阈值时输出告警。
type Alerter struct {
	cfg       *config.Config
	triggered bool
}

// NewAlerter 创建告警器；阈值为 0 表示不启用该项。
func NewAlerter(cfg *config.Config) *Alerter {
	return &Alerter{cfg: cfg}
}

// Triggered 是否曾触发过告警。
func (a *Alerter) Triggered() bool {
	return a.triggered
}

// Check 检查单条统计是否超阈值，超限时向 stderr 输出 [ALERT]。
func (a *Alerter) Check(stats models.ProcessStats) {
	if a.cfg.AlertCPUPercent > 0 && stats.CPUPercent >= a.cfg.AlertCPUPercent {
		a.fire("CPU", stats, fmt.Sprintf("%.1f%% >= %.1f%%", stats.CPUPercent, a.cfg.AlertCPUPercent))
	}
	if a.cfg.AlertMemGB > 0 && stats.MemoryGB >= a.cfg.AlertMemGB {
		a.fire("MEM", stats, fmt.Sprintf("%.3fGB >= %.3fGB", stats.MemoryGB, a.cfg.AlertMemGB))
	}
	if a.cfg.AlertGPUMemMB > 0 && stats.GPUMemoryMB >= a.cfg.AlertGPUMemMB {
		a.fire("GPU_MEM", stats, fmt.Sprintf("%.1fMB >= %.1fMB", stats.GPUMemoryMB, a.cfg.AlertGPUMemMB))
	}
}

// CheckContainer 检查容器 cgroup 内存限额使用率。
func (a *Alerter) CheckContainer(info models.ContainerInfo) {
	if a.cfg.AlertContainerMemPct <= 0 || !info.InContainer || info.MemLimitBytes == 0 {
		return
	}
	if info.MemUsagePct >= a.cfg.AlertContainerMemPct {
		a.triggered = true
		fmt.Fprintf(os.Stderr, "[ALERT] CTN_MEM 容器内存 %.1f%% >= %.1f%%（%.2f/%.2f GB）\n",
			info.MemUsagePct, a.cfg.AlertContainerMemPct, info.MemUsageGB(), info.MemLimitGB())
	}
}

func (a *Alerter) fire(kind string, stats models.ProcessStats, detail string) {
	a.triggered = true
	gpuTag := ""
	if stats.GPUIndex >= 0 {
		gpuTag = fmt.Sprintf(" GPU[%d]", stats.GPUIndex)
	}
	fmt.Fprintf(os.Stderr, "[ALERT] %s PID=%d%s %s %s\n", kind, stats.PID, gpuTag, stats.ProcessName, detail)
}

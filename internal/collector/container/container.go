// Package container 提供容器环境检测、cgroup 限额读取与进程过滤。
package container

import (
	"github.com/liubo/process-monitor/internal/models"
)

// Detect 检测当前运行环境并读取 cgroup 资源信息。
func Detect() models.ContainerInfo {
	info := detectPlatform()
	refreshCgroupStats(&info)
	return info
}

// FilterProcs 按监控范围过滤进程列表。
// scope: auto（容器内自动过滤）/ container（强制）/ host（不过滤）。
func FilterProcs(procs []models.ProcInfo, scope string, info models.ContainerInfo) []models.ProcInfo {
	if scope == "host" {
		return procs
	}
	if !info.InContainer {
		if scope == "container" {
			return procs
		}
		return procs
	}
	if scope == "auto" || scope == "container" {
		return filterByCgroup(procs, info.CgroupPath)
	}
	return procs
}

// RefreshStats 刷新 cgroup 内存/CPU 用量。
func RefreshStats(info *models.ContainerInfo) {
	refreshCgroupStats(info)
}

// ScopeActive 判断当前是否启用容器范围过滤。
func ScopeActive(scope string, info models.ContainerInfo) bool {
	if scope == "host" {
		return false
	}
	if scope == "container" {
		return info.InContainer
	}
	return info.InContainer // auto
}

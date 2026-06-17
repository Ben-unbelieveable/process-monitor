package monitor

import (
	"log"

	"github.com/liubo/process-monitor/internal/collector/container"
	"github.com/liubo/process-monitor/internal/config"
	"github.com/liubo/process-monitor/internal/models"
)

var runtimeContainer models.ContainerInfo

// initRuntimeEnv 检测并缓存容器运行环境。
func initRuntimeEnv() {
	runtimeContainer = container.Detect()
}

// getContainerInfo 返回当前容器环境信息。
func getContainerInfo() models.ContainerInfo {
	return runtimeContainer
}

// refreshContainerStats 刷新 cgroup 资源用量。
func refreshContainerStats() {
	container.RefreshStats(&runtimeContainer)
}

// printContainerInfo 输出容器环境与 cgroup 限额。
func printContainerInfo(cfg *config.Config) {
	info := getContainerInfo()
	log.Println("[Environment] 运行环境")
	if info.InContainer {
		log.Printf("  类型: 容器（%s）", info.Runtime)
		if info.PodNamespace != "" || info.PodName != "" {
			log.Printf("  Pod: %s/%s", info.PodNamespace, info.PodName)
		}
		if info.NodeName != "" {
			log.Printf("  节点: %s", info.NodeName)
		}
		if info.CgroupPath != "" {
			log.Printf("  cgroup: %s", info.CgroupPath)
		}
		if info.MemLimitBytes > 0 {
			log.Printf("  容器内存: %.2f/%.2f GB（%.0f%%）",
				info.MemUsageGB(), info.MemLimitGB(), info.MemUsagePct)
		} else if info.MemUsageBytes > 0 {
			log.Printf("  容器内存: %.2f GB（无限额）", info.MemUsageGB())
		}
		if info.CPUCoresLimit > 0 {
			log.Printf("  CPU 限额: %.1f 核", info.CPUCoresLimit)
		}
	} else {
		log.Println("  类型: 宿主机")
	}
	scope := cfg.EffectiveScope()
	switch scope {
	case "host":
		log.Println("  监控范围: 宿主机（-s host）")
	case "container":
		log.Println("  监控范围: 当前容器（-s container）")
	case "auto":
		if info.InContainer {
			log.Println("  监控范围: 当前容器（-s auto）")
		} else {
			log.Println("  监控范围: 宿主机（-s auto）")
		}
	}
	log.Println("--------------------------------------------------")
}

package monitor

import (
	"log"

	"github.com/liubo/process-monitor/internal/collector"
	"github.com/liubo/process-monitor/internal/config"
)

// printHardwareInfo 在监控开始前输出系统整体硬件配置。
func printHardwareInfo(cfg *config.Config) {
	hw := collector.GetSystemHardware(cfg.GPUType)
	log.Println("[Hardware] 系统硬件配置")
	if hw.Hostname != "" {
		log.Printf("  主机: %s", hw.Hostname)
	}
	log.Printf("  系统: %s/%s", hw.OS, hw.Arch)
	if hw.CPUModel != "" {
		log.Printf("  CPU: %s（%d 核 / %d 线程）", hw.CPUModel, hw.CPUCores, hw.CPUThreads)
	} else if hw.CPUCores > 0 {
		log.Printf("  CPU: %d 核 / %d 线程", hw.CPUCores, hw.CPUThreads)
	}
	if hw.MemTotalGB > 0 {
		log.Printf("  内存: %.1f GB", hw.MemTotalGB)
	}
	if len(hw.GPUs) == 0 {
		log.Println("  GPU: 未检测到")
	} else {
		for _, g := range hw.GPUs {
			if g.MemoryTotalMB > 0 {
				driver := g.DriverVersion
				if driver != "" {
					log.Printf("  GPU[%d]: %s | 显存 %.0f/%.0f MB | 利用率 %.0f%% | 驱动 %s",
						g.Index, g.Name, g.MemoryUsedMB, g.MemoryTotalMB, g.UtilizationPercent, driver)
				} else {
					log.Printf("  GPU[%d]: %s | 显存 %.0f/%.0f MB | 利用率 %.0f%%",
						g.Index, g.Name, g.MemoryUsedMB, g.MemoryTotalMB, g.UtilizationPercent)
				}
			} else {
				log.Printf("  GPU[%d]: %s", g.Index, g.Name)
			}
		}
	}
	log.Println("--------------------------------------------------")
}

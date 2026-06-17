// Package monitor 实现进程资源监控的主循环与数据采集编排。
package monitor

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/liubo/process-monitor/internal/collector"
	"github.com/liubo/process-monitor/internal/config"
	"github.com/liubo/process-monitor/internal/models"
	"github.com/liubo/process-monitor/internal/writer"
)

// Run 启动监控主循环，直到 stop 收到停止信号。
func Run(cfg *config.Config, stop <-chan struct{}) (alerted bool, err error) {
	initRuntimeEnv()
	printHardwareInfo(cfg)
	printContainerInfo(cfg)
	printStartupInfo(cfg)

	w, err := writer.New(cfg.OutputPath, cfg.OutputFormat)
	if err != nil {
		return false, fmt.Errorf("初始化输出: %w", err)
	}
	defer w.Close()

	alerter := NewAlerter(cfg)
	interval := time.Duration(cfg.Interval * float64(time.Second))

	for {
		select {
		case <-stop:
			log.Println("[Monitor] 监控已停止")
			return alerter.Triggered(), nil
		default:
		}

		statsList, err := collectStats(cfg)
		if err != nil {
			log.Printf("[ERROR] 采集异常: %v", err)
		} else {
			refreshContainerStats()
			alerter.CheckContainer(getContainerInfo())
			if len(statsList) > 0 {
				for _, s := range statsList {
					alerter.Check(s)
				}
				if err := w.WriteRows(statsList); err != nil {
					log.Printf("[ERROR] 输出异常: %v", err)
				} else if cfg.OutputToStdout() {
					log.Printf("[%s] 采集到 %d 条记录", time.Now().Format("15:04:05"), len(statsList))
				} else {
					log.Printf("[%s] 采集到 %d 条记录，已写入 %s", time.Now().Format("15:04:05"), len(statsList), cfg.OutputPath)
				}
				if cfg.AlertExit && alerter.Triggered() {
					log.Println("[Monitor] 告警触发，退出（--alert-exit）")
					return true, nil
				}
			} else {
				log.Printf("[%s] 无符合条件的进程", time.Now().Format("15:04:05"))
			}
		}

		select {
		case <-stop:
			log.Println("[Monitor] 监控已停止")
			return alerter.Triggered(), nil
		case <-time.After(interval):
		}
	}
}

func printStartupInfo(cfg *config.Config) {
	log.Println("[Monitor] 启动资源监控")
	if cfg.OutputToStdout() {
		log.Printf("  输出目标: 屏幕（%s）", cfg.OutputFormat)
	} else {
		log.Printf("  输出目标: %s（%s）", cfg.OutputPath, cfg.OutputFormat)
	}
	log.Printf("  监控间隔: %.0fs", cfg.Interval)
	log.Printf("  GPU 类型: %s", cfg.GPUType)
	if len(cfg.GPUIndexes) > 0 {
		log.Printf("  GPU 过滤: %v", cfg.GPUIndexes)
	}
	if cfg.Tree {
		log.Println("  进程树: 开启（汇总子孙进程）")
	}

	switch {
	case len(cfg.TargetPIDs) > 0:
		log.Printf("  指定 PID: %v", cfg.TargetPIDs)
	case len(cfg.TargetNames) > 0:
		log.Printf("  进程名匹配: %v", cfg.TargetNames)
	case len(cfg.TargetCmdlines) > 0:
		log.Printf("  命令行匹配: %v", cfg.TargetCmdlines)
	default:
		log.Printf("  自动模式: 内存 > %.2fGB 的进程", cfg.MemoryThresholdGB)
	}

	if cfg.AlertCPUPercent > 0 || cfg.AlertMemGB > 0 || cfg.AlertGPUMemMB > 0 || cfg.AlertContainerMemPct > 0 {
		var parts []string
		if cfg.AlertCPUPercent > 0 {
			parts = append(parts, fmt.Sprintf("CPU>=%.0f%%", cfg.AlertCPUPercent))
		}
		if cfg.AlertMemGB > 0 {
			parts = append(parts, fmt.Sprintf("MEM>=%.2fGB", cfg.AlertMemGB))
		}
		if cfg.AlertGPUMemMB > 0 {
			parts = append(parts, fmt.Sprintf("GPU_MEM>=%.0fMB", cfg.AlertGPUMemMB))
		}
		if cfg.AlertContainerMemPct > 0 {
			parts = append(parts, fmt.Sprintf("CTN_MEM>=%.0f%%", cfg.AlertContainerMemPct))
		}
		log.Printf("  告警阈值: %s", strings.Join(parts, " "))
		if cfg.AlertExit {
			log.Println("  告警退出: 开启")
		}
	}
	log.Println("--------------------------------------------------")
}

func collectStats(cfg *config.Config) ([]models.ProcessStats, error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	groups := resolveTargets(cfg)
	snap := resolveGPUSnapshot(cfg)

	var results []models.ProcessStats
	for _, g := range groups {
		if len(g.PIDs) == 1 && !cfg.Tree {
			stats := collector.GetProcessStats(g.PIDs[0], g.Name)
			if stats == nil {
				continue
			}
			procGPUs := collector.CollectProcGPUs(g.PIDs[0], cfg.GPUType)
			rows := buildProcessRows(timestamp, g.RootPID, g.Name, stats.CPUPercent, stats.MemoryGB, procGPUs, snap, cfg)
			results = append(results, rows...)
			continue
		}

		cpu, mem, name := collector.AggregateStats(g.PIDs, g.Name)
		procGPUs := collector.CollectTreeGPUs(g.PIDs, cfg.GPUType)
		rows := buildProcessRows(timestamp, g.RootPID, name, cpu, mem, procGPUs, snap, cfg)
		results = append(results, rows...)
	}
	return results, nil
}

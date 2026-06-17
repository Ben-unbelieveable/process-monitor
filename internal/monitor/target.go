package monitor

import (
	"github.com/liubo/process-monitor/internal/collector"
	"github.com/liubo/process-monitor/internal/collector/container"
	"github.com/liubo/process-monitor/internal/config"
	"github.com/liubo/process-monitor/internal/models"
)

// TargetGroup 表示一个监控目标（可选含子孙进程）。
type TargetGroup struct {
	RootPID int32
	Name    string
	PIDs    []int32
}

// resolveTargets 根据配置解析监控目标列表。
func resolveTargets(cfg *config.Config) []TargetGroup {
	allProcs := collector.GetAllProcesses()
	allProcs = container.FilterProcs(allProcs, cfg.EffectiveScope(), getContainerInfo())
	parentOf := collector.BuildParentMap()
	childrenOf := collector.BuildChildMap()

	var matches []models.ProcInfo
	switch {
	case len(cfg.TargetPIDs) > 0:
		for _, pid := range cfg.TargetPIDs {
			matches = append(matches, collector.LookupProcByPID(allProcs, pid))
		}
	case len(cfg.TargetNames) > 0:
		matches = collector.FilterByNames(allProcs, cfg.TargetNames)
	case len(cfg.TargetCmdlines) > 0:
		matches = collector.FilterByCmdline(allProcs, cfg.TargetCmdlines)
	default:
		matches = collector.FilterByMemory(allProcs, cfg.MemoryThresholdGB)
	}

	if cfg.Tree && len(cfg.TargetPIDs)+len(cfg.TargetNames)+len(cfg.TargetCmdlines) > 0 {
		matches = collector.FilterTopLevel(matches, parentOf)
	}

	var groups []TargetGroup
	for _, m := range matches {
		pids := []int32{m.PID}
		if cfg.Tree {
			pids = collector.TreePIDs(m.PID, childrenOf)
		}
		groups = append(groups, TargetGroup{
			RootPID: m.PID,
			Name:    m.Name,
			PIDs:    pids,
		})
	}
	return groups
}

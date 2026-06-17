//go:build !linux

package container

import (
	"github.com/liubo/process-monitor/internal/models"
)

func detectPlatform() models.ContainerInfo {
	return models.ContainerInfo{Runtime: "none"}
}

func filterByCgroup(procs []models.ProcInfo, _ string) []models.ProcInfo {
	return procs
}

func refreshCgroupStats(_ *models.ContainerInfo) {}

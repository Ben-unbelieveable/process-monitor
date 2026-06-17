package monitor

import (
	"log"

	"github.com/liubo/process-monitor/internal/version"
)

// printVersionInfo 向日志输出版本与仓库信息（用于文件输出模式）。
func printVersionInfo() {
	for _, line := range version.HeaderLines() {
		log.Println(line)
	}
}

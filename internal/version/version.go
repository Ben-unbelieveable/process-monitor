// Package version 保存构建时注入的版本与仓库信息。
package version

import "fmt"

// 以下变量在编译时通过 -ldflags -X 注入。
var (
	Version    = "dev"
	BuildDate  = "unknown"
	Repository = "https://github.com/Ben-unbelieveable/process-monitor"
)

// Banner 返回单行版本摘要，用于帮助文档与日志抬头。
func Banner() string {
	return fmt.Sprintf("%s | build %s | %s", Version, BuildDate, Repository)
}

// HeaderLines 返回写入统计文件/屏幕前的元信息行。
func HeaderLines() []string {
	return []string{
		fmt.Sprintf("# process-monitor %s", Version),
		fmt.Sprintf("# build: %s", BuildDate),
		fmt.Sprintf("# repository: %s", Repository),
	}
}

// MetaJSON 返回 JSON Lines 格式的元信息对象。
func MetaJSON() string {
	return fmt.Sprintf(
		`{"_meta":{"version":%q,"build":%q,"repository":%q}}`,
		Version, BuildDate, Repository,
	)
}

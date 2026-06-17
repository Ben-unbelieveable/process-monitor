package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// printUsage 按功能分组输出帮助信息。
func printUsage(fs *flag.FlagSet) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Process Monitor — 进程资源监控工具\n\n")
	fmt.Fprintf(w, "用法:\n  %s [选项]\n\n", fs.Name())

	writeSection(w, "监控目标", "（-p / -n / -a 三选一；均未指定则自动模式）", [][]string{
		{"-p, --pid", "<pid>", "监控指定 PID（可重复）"},
		{"-n, --name", "<关键词>", "按进程名匹配（可重复）"},
		{"-a, --args", "<关键词>", "按命令行匹配（可重复）"},
		{"--tree", "", "汇总子孙进程 CPU/内存/GPU"},
		{"-m, --memory-threshold", "<GB>", "自动模式内存阈值（默认 1.0）"},
		{"-s, --scope", "<范围>", "监控范围 auto|host|container（默认 auto）"},
		{"--gpu-type", "<类型>", "GPU 类型 nvidia|apple|both（默认 both）"},
		{"--gpu-index", "<编号>", "只监控指定 GPU（可重复，默认全部）"},
	})

	writeSection(w, "输出与采集", "", [][]string{
		{"-i, --interval", "<秒>", "采集间隔（默认 30）"},
		{"-o, --output", "<路径>", "输出文件（默认屏幕）"},
		{"-f, --format", "<格式>", "text|csv|tsv|json（默认 text）"},
		{"-c, --config", "<文件>", "可选 YAML 配置文件"},
	})

	writeSection(w, "告警阈值", "（0 表示关闭该项）", [][]string{
		{"--alert-cpu", "<%>", "进程 CPU 使用率"},
		{"--alert-mem-gb", "<GB>", "进程内存"},
		{"--alert-gpu-mem-mb", "<MB>", "进程 GPU 显存"},
		{"--alert-container-mem-pct", "<%>", "容器内存限额使用率"},
		{"--alert-exit", "", "触发告警后立即退出（退出码 2）"},
	})

	w.Flush()
	fmt.Println("\n示例:")
	fmt.Println("  monitor -a train.py --tree -i 10")
	fmt.Println("  monitor -p 12345 -s container --alert-mem-gb 32")
	fmt.Println("  monitor -n python -o stats.csv -f csv --alert-cpu 90")
}

func writeSection(w io.Writer, title, subtitle string, rows [][]string) {
	fmt.Fprintf(w, "%s", title)
	if subtitle != "" {
		fmt.Fprintf(w, " %s", subtitle)
	}
	fmt.Fprintf(w, ":\n")
	for _, row := range rows {
		flags, arg, desc := row[0], row[1], row[2]
		if arg != "" {
			fmt.Fprintf(w, "  %-28s %s\t%s\n", flags, arg, desc)
		} else {
			fmt.Fprintf(w, "  %-28s\t%s\n", flags, desc)
		}
	}
	fmt.Fprintln(w)
}

// registerFlags 注册全部命令行参数（简写与全拼共用变量，帮助由 printUsage 统一输出）。
func registerFlags(fs *flag.FlagSet, args *cliArgs) {
	regDualInt32(fs, "p", "pid", &args.pids)
	regDualString(fs, "n", "name", &args.names)
	regDualString(fs, "a", "args", &args.cmdlines)
	regDualFloat(fs, "i", "interval", &args.interval, 30)
	regDualStr(fs, "o", "output", &args.output, "")
	regDualStr(fs, "f", "format", &args.format, "text")
	regDualStr(fs, "c", "config", &args.configPath, "")
	regDualFloat(fs, "m", "memory-threshold", &args.memThreshold, 1.0)
	regDualStr(fs, "s", "scope", &args.scope, "")

	fs.BoolVar(&args.tree, "tree", false, "")
	fs.StringVar(&args.gpuType, "gpu-type", "both", "")
	fs.Var(&args.gpuIndexes, "gpu-index", "")
	fs.Float64Var(&args.alertCPU, "alert-cpu", 0, "")
	fs.Float64Var(&args.alertMem, "alert-mem-gb", 0, "")
	fs.Float64Var(&args.alertGPU, "alert-gpu-mem-mb", 0, "")
	fs.Float64Var(&args.alertCtnMem, "alert-container-mem-pct", 0, "")
	fs.BoolVar(&args.alertExit, "alert-exit", false, "")
	// 兼容旧参数，不出现在分组帮助中
	fs.BoolVar(&args.hostScope, "host", false, "")
	fs.BoolVar(&args.ctnOnly, "container-only", false, "")
}

func regDualInt32(fs *flag.FlagSet, short, long string, dest *int32Slice) {
	fs.Var(dest, short, "")
	fs.Var(dest, long, "")
}

func regDualString(fs *flag.FlagSet, short, long string, dest *stringSlice) {
	fs.Var(dest, short, "")
	fs.Var(dest, long, "")
}

func regDualStr(fs *flag.FlagSet, short, long string, dest *string, def string) {
	fs.StringVar(dest, short, def, "")
	fs.StringVar(dest, long, def, "")
}

func regDualFloat(fs *flag.FlagSet, short, long string, dest *float64, def float64) {
	fs.Float64Var(dest, short, def, "")
	fs.Float64Var(dest, long, def, "")
}

// cliArgs 保存命令行原始解析结果。
type cliArgs struct {
	configPath   string
	pids         int32Slice
	names        stringSlice
	cmdlines     stringSlice
	interval     float64
	output       string
	format       string
	gpuType      string
	memThreshold float64
	scope        string
	tree         bool
	alertCPU     float64
	alertMem     float64
	alertGPU     float64
	alertCtnMem  float64
	alertExit    bool
	hostScope    bool
	ctnOnly      bool
	gpuIndexes   intSlice
}

func (a *cliArgs) apply(setFlags map[string]bool, cfg *Config) error {
	if flagSet(setFlags, "p", "pid") {
		cfg.TargetPIDs = a.pids
	}
	if flagSet(setFlags, "n", "name") {
		cfg.TargetNames = a.names
	}
	if flagSet(setFlags, "a", "args") {
		cfg.TargetCmdlines = a.cmdlines
	}
	if flagSet(setFlags, "i", "interval") {
		cfg.Interval = a.interval
	}
	if flagSet(setFlags, "o", "output") {
		cfg.OutputPath = a.output
	}
	if flagSet(setFlags, "f", "format") {
		cfg.OutputFormat = a.format
	}
	if flagSet(setFlags, "c", "config") {
		// configPath handled before apply
	}
	if flagSet(setFlags, "m", "memory-threshold") {
		cfg.MemoryThresholdGB = a.memThreshold
	}
	if flagSet(setFlags, "s", "scope") {
		s := strings.ToLower(strings.TrimSpace(a.scope))
		switch s {
		case "auto", "host", "container":
			cfg.Scope = s
		default:
			return fmt.Errorf("scope 无效: %q（可选 auto|host|container）", a.scope)
		}
	}
	if setFlags["gpu-type"] {
		cfg.GPUType = a.gpuType
	}
	if setFlags["gpu-index"] {
		cfg.GPUIndexes = a.gpuIndexes
	}
	if setFlags["tree"] {
		cfg.Tree = a.tree
	}
	if setFlags["alert-cpu"] {
		cfg.AlertCPUPercent = a.alertCPU
	}
	if setFlags["alert-mem-gb"] {
		cfg.AlertMemGB = a.alertMem
	}
	if setFlags["alert-gpu-mem-mb"] {
		cfg.AlertGPUMemMB = a.alertGPU
	}
	if setFlags["alert-container-mem-pct"] {
		cfg.AlertContainerMemPct = a.alertCtnMem
	}
	if setFlags["alert-exit"] {
		cfg.AlertExit = a.alertExit
	}
	if setFlags["host"] {
		cfg.HostScope = a.hostScope
	}
	if setFlags["container-only"] {
		cfg.ContainerOnly = a.ctnOnly
	}
	return nil
}

func flagSet(setFlags map[string]bool, names ...string) bool {
	for _, n := range names {
		if setFlags[n] {
			return true
		}
	}
	return false
}

func visitedFlags(fs *flag.FlagSet) map[string]bool {
	setFlags := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })
	return setFlags
}

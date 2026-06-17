// Package config 负责解析命令行参数，并可选地从 YAML 文件加载配置。
package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 保存监控运行时配置。
type Config struct {
	Interval          float64  `yaml:"interval"`
	OutputPath        string   `yaml:"output"`
	OutputFormat      string   `yaml:"output_format"`
	GPUType           string   `yaml:"gpu_type"`
	GPUIndexes        []int    `yaml:"gpu_indexes"`
	TargetPIDs        []int32  `yaml:"target_pids"`
	TargetNames       []string `yaml:"target_names"`
	TargetCmdlines    []string `yaml:"target_cmdlines"`
	MemoryThresholdGB float64  `yaml:"memory_threshold_gb"`
	Tree              bool     `yaml:"tree"`
	AlertCPUPercent   float64  `yaml:"alert_cpu_percent"`
	AlertMemGB        float64  `yaml:"alert_mem_gb"`
	AlertGPUMemMB          float64  `yaml:"alert_gpu_mem_mb"`
	AlertContainerMemPct   float64  `yaml:"alert_container_mem_pct"`
	AlertExit              bool     `yaml:"alert_exit"`
	Scope                  string   `yaml:"scope"` // auto / host / container
	HostScope              bool     `yaml:"host_scope"`       // 兼容旧配置
	ContainerOnly          bool     `yaml:"container_only"`   // 兼容旧配置
}

// Default 返回 CLI 默认值。
func Default() *Config {
	return &Config{
		Interval:          30,
		OutputPath:        "",
		OutputFormat:      "text",
		GPUType:           "both",
		TargetPIDs:        []int32{},
		TargetNames:       []string{},
		TargetCmdlines:    []string{},
		MemoryThresholdGB: 1.0,
		GPUIndexes:        []int{},
	}
}

// GPUIndexAllowed 判断指定 GPU 索引是否在过滤范围内（空列表表示全部）。
func (c *Config) GPUIndexAllowed(idx int) bool {
	if len(c.GPUIndexes) == 0 {
		return true
	}
	for _, i := range c.GPUIndexes {
		if i == idx {
			return true
		}
	}
	return false
}

// EffectiveScope 返回进程监控范围：auto / host / container。
func (c *Config) EffectiveScope() string {
	if c.Scope != "" {
		return c.Scope
	}
	if c.HostScope {
		return "host"
	}
	if c.ContainerOnly {
		return "container"
	}
	return "auto"
}

// OutputToStdout 是否输出到屏幕。
func (c *Config) OutputToStdout() bool {
	return c.OutputPath == "" || c.OutputPath == "-"
}

// Normalize 校验并规范化配置。
func (c *Config) Normalize() error {
	if c.Interval <= 0 {
		c.Interval = 30
	}
	if c.OutputFormat == "" {
		if c.OutputToStdout() {
			c.OutputFormat = "text"
		} else {
			c.OutputFormat = "csv"
		}
	}
	if c.GPUType == "" {
		c.GPUType = "both"
	}
	if c.TargetPIDs == nil {
		c.TargetPIDs = []int32{}
	}
	if c.TargetNames == nil {
		c.TargetNames = []string{}
	}
	if c.TargetCmdlines == nil {
		c.TargetCmdlines = []string{}
	}
	if c.GPUIndexes == nil {
		c.GPUIndexes = []int{}
	}
	if c.MemoryThresholdGB <= 0 {
		c.MemoryThresholdGB = 1.0
	}

	switch c.OutputFormat {
	case "text", "csv", "tsv", "json":
	default:
		return fmt.Errorf("不支持的输出格式: %s（可选 text/csv/tsv/json）", c.OutputFormat)
	}
	switch c.GPUType {
	case "nvidia", "apple", "both":
	default:
		return fmt.Errorf("不支持的 GPU 类型: %s（可选 nvidia/apple/both）", c.GPUType)
	}

	if !c.OutputToStdout() {
		extMap := map[string]string{"csv": ".csv", "tsv": ".tsv", "json": ".json", "text": ".log"}
		ext, ok := extMap[c.OutputFormat]
		if !ok {
			ext = ".csv"
		}
		base := c.OutputPath
		if idx := strings.LastIndex(base, "."); idx >= 0 {
			base = base[:idx]
		}
		if !strings.HasSuffix(c.OutputPath, ext) {
			c.OutputPath = base + ext
		}
	}
	return nil
}

// Validate 校验目标筛选互斥等业务规则。
func (c *Config) Validate() error {
	modes := 0
	if len(c.TargetPIDs) > 0 {
		modes++
	}
	if len(c.TargetNames) > 0 {
		modes++
	}
	if len(c.TargetCmdlines) > 0 {
		modes++
	}
	if modes > 1 {
		return errors.New("不能同时指定 -p、-n、-a，请三选一")
	}
	return nil
}

// Load 从 YAML 文件加载配置（可选）。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置: %w", err)
	}
	if cfg.OutputPath == "" {
		var legacy struct {
			OutputDir  string `yaml:"output_dir"`
			OutputFile string `yaml:"output_file"`
		}
		_ = yaml.Unmarshal(data, &legacy)
		if legacy.OutputFile != "" {
			if legacy.OutputDir != "" {
				cfg.OutputPath = filepath.Join(legacy.OutputDir, legacy.OutputFile)
			} else {
				cfg.OutputPath = legacy.OutputFile
			}
		}
	}
	return cfg, nil
}

type int32Slice []int32

func (s *int32Slice) String() string { return fmt.Sprint(*s) }

func (s *int32Slice) Set(val string) error {
	var v int
	if _, err := fmt.Sscanf(val, "%d", &v); err != nil {
		return err
	}
	*s = append(*s, int32(v))
	return nil
}

type intSlice []int

func (s *intSlice) String() string { return fmt.Sprint(*s) }

func (s *intSlice) Set(val string) error {
	var v int
	if _, err := fmt.Sscanf(val, "%d", &v); err != nil {
		return err
	}
	*s = append(*s, v)
	return nil
}

type stringSlice []string

func (s *stringSlice) String() string { return fmt.Sprint(*s) }

func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

// Parse 解析命令行参数，可选加载配置文件，返回最终配置。
func Parse() (*Config, error) {
	args := &cliArgs{}
	fs := flag.NewFlagSet("monitor", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() { printUsage(fs) }
	registerFlags(fs, args)

	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			fs.Usage()
			os.Exit(0)
		}
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	setFlags := visitedFlags(fs)
	cfg := Default()

	if args.configPath != "" || setFlags["c"] || setFlags["config"] {
		path := args.configPath
		if path == "" {
			return nil, fmt.Errorf("请通过 -c / --config 指定配置文件路径")
		}
		fileCfg, err := Load(path)
		if err != nil {
			return nil, err
		}
		mergeConfig(cfg, fileCfg)
	}

	if err := args.apply(setFlags, cfg); err != nil {
		return nil, err
	}

	if err := cfg.Normalize(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func mergeConfig(dst, src *Config) {
	if src.Interval > 0 {
		dst.Interval = src.Interval
	}
	if src.OutputPath != "" {
		dst.OutputPath = src.OutputPath
	}
	if src.OutputFormat != "" {
		dst.OutputFormat = src.OutputFormat
	}
	if src.GPUType != "" {
		dst.GPUType = src.GPUType
	}
	if len(src.GPUIndexes) > 0 {
		dst.GPUIndexes = src.GPUIndexes
	}
	if len(src.TargetPIDs) > 0 {
		dst.TargetPIDs = src.TargetPIDs
	}
	if len(src.TargetNames) > 0 {
		dst.TargetNames = src.TargetNames
	}
	if len(src.TargetCmdlines) > 0 {
		dst.TargetCmdlines = src.TargetCmdlines
	}
	if src.MemoryThresholdGB > 0 {
		dst.MemoryThresholdGB = src.MemoryThresholdGB
	}
	if src.Tree {
		dst.Tree = true
	}
	if src.AlertCPUPercent > 0 {
		dst.AlertCPUPercent = src.AlertCPUPercent
	}
	if src.AlertMemGB > 0 {
		dst.AlertMemGB = src.AlertMemGB
	}
	if src.AlertGPUMemMB > 0 {
		dst.AlertGPUMemMB = src.AlertGPUMemMB
	}
	if src.AlertContainerMemPct > 0 {
		dst.AlertContainerMemPct = src.AlertContainerMemPct
	}
	if src.AlertExit {
		dst.AlertExit = true
	}
	if src.Scope != "" {
		dst.Scope = src.Scope
	}
	if src.HostScope {
		dst.HostScope = true
	}
	if src.ContainerOnly {
		dst.ContainerOnly = true
	}
}

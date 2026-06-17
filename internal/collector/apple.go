package collector

import (
	"encoding/json"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// IsAppleGPUAvailable 检测当前是否为 Apple Silicon（macOS ARM64）环境。
func IsAppleGPUAvailable() bool {
	return runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"
}

// AppleGPUInfo 表示 Apple Silicon GPU 信息。
type AppleGPUInfo struct {
	GPUName        string
	GPUMemoryMB    float64
	MemoryTotalMB  float64
	GPUUtilization float64
}

// GetGPUInfoApple 通过 system_profiler 获取 Apple Silicon GPU 信息。
func GetGPUInfoApple() []AppleGPUInfo {
	if !IsAppleGPUAvailable() {
		return nil
	}

	out, err := exec.Command("system_profiler", "SPDisplaysDataType", "-json").Output()
	if err != nil {
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(out, &data); err != nil {
		return nil
	}

	displays, ok := data["SPDisplaysDataType"].([]interface{})
	if !ok {
		return nil
	}

	var results []AppleGPUInfo
	for _, item := range displays {
		disp, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		gpuName := "Apple Silicon GPU"
		if name, ok := disp["_name"]; ok {
			gpuName = toString(name)
		} else if cores, ok := disp["sppci_cores"]; ok {
			gpuName = "Apple GPU (" + toString(cores) + " cores)"
		}
		results = append(results, AppleGPUInfo{
			GPUName:        gpuName,
			GPUMemoryMB:    0,
			GPUUtilization: 0,
		})
	}
	return results
}

// GetGPUMapApple 返回 Apple GPU 映射（通常为 index 0）。
func GetGPUMapApple() map[int]AppleGPUInfo {
	gpus := GetGPUInfoApple()
	m := make(map[int]AppleGPUInfo, len(gpus))
	for i, g := range gpus {
		m[i] = g
	}
	return m
}

// ProcessMetalMemory 获取指定进程在 Apple GPU 上的近似内存使用（MB）。
func ProcessMetalMemory(pid int32) *AppleGPUInfo {
	out, err := exec.Command("ps", "-o", "pid,rss", "-p", strconv.Itoa(int(pid))).Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return nil
	}
	parts := strings.Fields(lines[1])
	if len(parts) < 2 {
		return nil
	}
	rssKB, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil
	}
	return &AppleGPUInfo{GPUMemoryMB: rssKB / 1024}
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}

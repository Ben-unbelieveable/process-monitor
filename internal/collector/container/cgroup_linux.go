//go:build linux

package container

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/liubo/process-monitor/internal/models"
)

const cgroupV2Mount = "/sys/fs/cgroup"

func detectPlatform() models.ContainerInfo {
	info := models.ContainerInfo{
		PodName:      readEnv("POD_NAME"),
		PodNamespace: readEnv("POD_NAMESPACE"),
		NodeName:     readEnv("NODE_NAME"),
	}
	if info.PodName == "" {
		info.PodName = readEnv("HOSTNAME")
	}

	if fileExists("/.dockerenv") {
		info.InContainer = true
		info.Runtime = "docker"
	}
	if fileExists("/run/.containerenv") {
		info.InContainer = true
		info.Runtime = "podman"
	}

	cgroupPath, content := readSelfCgroup()
	info.CgroupPath = cgroupPath
	if cgroupPath != "" && isContainerCgroupPath(cgroupPath) {
		info.InContainer = true
		if info.Runtime == "" {
			info.Runtime = detectRuntimeFromCgroup(content, cgroupPath)
		}
	}
	if !info.InContainer {
		info.Runtime = "none"
	}
	return info
}

func detectRuntimeFromCgroup(content, path string) string {
	lower := strings.ToLower(content + path)
	switch {
	case strings.Contains(lower, "kubepods"), strings.Contains(lower, "kubelet"):
		return "kubernetes"
	case strings.Contains(lower, "docker"):
		return "docker"
	case strings.Contains(lower, "containerd"), strings.Contains(lower, "cri-containerd"):
		return "containerd"
	case strings.Contains(lower, "libpod"), strings.Contains(lower, "podman"):
		return "podman"
	case strings.Contains(lower, "lxc"):
		return "lxc"
	default:
		return "container"
	}
}

func isContainerCgroupPath(path string) bool {
	if path == "/" || path == "" {
		return false
	}
	lower := strings.ToLower(path)
	markers := []string{"docker", "kubepods", "kube", "containerd", "cri-", "libpod", "lxc", "machine.slice"}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

func readSelfCgroup() (unifiedPath, rawContent string) {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return "", ""
	}
	rawContent = string(data)
	// cgroup v2 统一层级
	for _, line := range strings.Split(rawContent, "\n") {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		if parts[0] == "0" && parts[1] == "" {
			return parts[2], rawContent
		}
	}
	// cgroup v1 回退：取 memory 或 name=systemd
	for _, line := range strings.Split(rawContent, "\n") {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		if parts[1] == "memory" || parts[1] == "name=systemd" {
			return parts[2], rawContent
		}
	}
	return "", rawContent
}

func filterByCgroup(procs []models.ProcInfo, selfPath string) []models.ProcInfo {
	if selfPath == "" {
		return procs
	}
	key := cgroupMatchKey(selfPath)
	var matched []models.ProcInfo
	for _, p := range procs {
		if pidInSameCgroup(p.PID, selfPath, key) {
			matched = append(matched, p)
		}
	}
	return matched
}

func cgroupMatchKey(path string) string {
	// 优先用完整路径前缀；对 k8s/docker 提取特征段
	for _, seg := range strings.Split(path, "/") {
		if strings.HasPrefix(seg, "pod") && len(seg) > 10 {
			return seg
		}
		if strings.HasPrefix(seg, "docker-") || strings.HasPrefix(seg, "cri-containerd-") {
			return seg
		}
	}
	return path
}

func pidInSameCgroup(pid int32, selfPath, key string) bool {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return false
	}
	content := string(data)
	if strings.Contains(content, key) {
		return true
	}
	for _, line := range strings.Split(content, "\n") {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		p := parts[2]
		if p == selfPath || strings.HasPrefix(p, selfPath+"/") || strings.HasPrefix(selfPath, p+"/") {
			return true
		}
	}
	return false
}

func refreshCgroupStats(info *models.ContainerInfo) {
	if !info.InContainer || info.CgroupPath == "" {
		return
	}
	root := cgroupV2Root(info.CgroupPath)
	if root == "" {
		return
	}
	info.MemUsageBytes = readUintFile(filepath.Join(root, "memory.current"))
	if limit := readCgroupValue(root, "memory.max"); limit != "" && limit != "max" {
		if v, err := strconv.ParseUint(limit, 10, 64); err == nil {
			info.MemLimitBytes = v
		}
	}
	if info.MemLimitBytes > 0 && info.MemUsageBytes > 0 {
		info.MemUsagePct = float64(info.MemUsageBytes) / float64(info.MemLimitBytes) * 100
	} else {
		info.MemUsagePct = -1
	}
	info.CPUCoresLimit = readCPUCoresLimit(root)
}

func cgroupV2Root(cgroupPath string) string {
	// 优先 cgroup v2 路径
	v2 := filepath.Join(cgroupV2Mount, strings.TrimPrefix(cgroupPath, "/"))
	if fileExists(filepath.Join(v2, "memory.current")) {
		return v2
	}
	// cgroup v1 回退
	v1mem := filepath.Join(cgroupV2Mount, "memory", strings.TrimPrefix(cgroupPath, "/"))
	if fileExists(filepath.Join(v1mem, "memory.usage_in_bytes")) {
		return v1mem
	}
	v1cpu := filepath.Join(cgroupV2Mount, "cpu,cpuacct", strings.TrimPrefix(cgroupPath, "/"))
	if fileExists(filepath.Join(v1cpu, "cpu.cfs_quota_us")) {
		// memory 可能在独立路径，仍尝试 v2
		return v2
	}
	return v2
}

func readCPUCoresLimit(root string) float64 {
	if val := readCgroupValue(root, "cpu.max"); val != "" && val != "max" {
		fields := strings.Fields(val)
		if len(fields) == 2 {
			quota, _ := strconv.ParseFloat(fields[0], 64)
			period, _ := strconv.ParseFloat(fields[1], 64)
			if period > 0 && quota > 0 {
				return quota / period
			}
		}
	}
	// cgroup v1
	quota := readIntFile(filepath.Join(root, "cpu.cfs_quota_us"))
	period := readIntFile(filepath.Join(root, "cpu.cfs_period_us"))
	if quota > 0 && period > 0 {
		return float64(quota) / float64(period)
	}
	return 0
}

func readCgroupValue(root, name string) string {
	data, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readUintFile(path string) uint64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	v, _ := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	return v
}

func readIntFile(path string) int64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	v, _ := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	return v
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

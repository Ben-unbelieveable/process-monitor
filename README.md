# Process Monitor

跨平台进程资源监控工具，用于 ML/DL 训练和预测过程中的 CPU/GPU/内存追踪。

## 构建

```bash
make build
go build -o bin/monitor ./cmd/monitor
```

## 帮助

```bash
./bin/monitor -h
```

参数按功能分为三组：**监控目标**、**输出与采集**、**告警阈值**。简写与全拼等价，例如 `-a` 与 `--args` 相同。

## 命令行参数摘要

### 监控目标（-p / -n / -a 三选一）

| 参数 | 说明 |
|------|------|
| `-p, --pid <pid>` | 监控指定 PID（可重复） |
| `-n, --name <关键词>` | 按进程名匹配（可重复） |
| `-a, --args <关键词>` | 按命令行匹配（可重复） |
| `--tree` | 汇总子孙进程资源 |
| `-m, --memory-threshold <GB>` | 自动模式内存阈值（默认 1.0） |
| `-s, --scope <范围>` | `auto` / `host` / `container`（默认 auto） |
| `--gpu-type` / `--gpu-index` | GPU 类型与过滤 |

### 输出与采集

| 参数 | 说明 |
|------|------|
| `-i, --interval <秒>` | 采集间隔（默认 30） |
| `-o, --output <路径>` | 输出文件（默认屏幕） |
| `-f, --format <格式>` | text / csv / tsv / json |
| `-c, --config <文件>` | 可选 YAML 配置 |

### 告警阈值（0 关闭）

| 参数 | 说明 |
|------|------|
| `--alert-cpu` | 进程 CPU % |
| `--alert-mem-gb` | 进程内存 GB |
| `--alert-gpu-mem-mb` | GPU 显存 MB |
| `--alert-container-mem-pct` | 容器内存限额 % |
| `--alert-exit` | 告警后立即退出 |

## 使用示例

```bash
./bin/monitor-darwin-arm64 -a train.py --tree -i 10
./bin/monitor-darwin-arm64 -p 12345 -s container --alert-mem-gb 32
./bin/monitor-linux-amd64 -n python -o stats.csv -f csv --alert-cpu 90
```

## 容器感知（Linux）

容器内默认 `-s auto` 仅监控本容器进程；`-s host` 监控宿主机全部进程。

```bash
./bin/monitor-linux-amd64 -s container --alert-container-mem-pct 90
```

## 注意事项

- 容器感知完整支持 **Linux**；macOS 显示宿主机
- NVIDIA GPU 需要 `nvidia-smi`

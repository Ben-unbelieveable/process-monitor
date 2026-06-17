// Package writer 提供屏幕 / 文件输出，支持 text、CSV、TSV、JSON Lines 格式。
package writer

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/liubo/process-monitor/internal/models"
	"github.com/liubo/process-monitor/internal/version"
)

// StatsWriter 将 ProcessStats 写入屏幕或文件。
type StatsWriter struct {
	out        io.Writer
	format     string
	file       *os.File
	csvWriter  *csv.Writer
	firstWrite bool
}

// New 创建 StatsWriter。output 为空或 "-" 时输出到 stdout。
func New(output, format string) (*StatsWriter, error) {
	w := &StatsWriter{
		format:     format,
		firstWrite: true,
	}
	if output == "" || output == "-" {
		w.out = os.Stdout
		return w, nil
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return nil, fmt.Errorf("创建输出目录: %w", err)
	}
	f, err := os.Create(output)
	if err != nil {
		return nil, fmt.Errorf("创建输出文件: %w", err)
	}
	w.file = f
	w.out = f
	return w, nil
}

func (w *StatsWriter) init() error {
	if !w.firstWrite {
		return nil
	}
	if err := w.writeMetaHeader(); err != nil {
		return err
	}
	switch w.format {
	case "json":
		// JSON Lines：每行一个 JSON 对象
	case "text":
		if _, err := io.WriteString(w.out, textTableHeader()+"\n"); err != nil {
			return fmt.Errorf("写入表头: %w", err)
		}
		if _, err := io.WriteString(w.out, textTableSeparator()+"\n"); err != nil {
			return fmt.Errorf("写入分隔线: %w", err)
		}
	case "csv", "tsv":
		delimiter := ','
		if w.format == "tsv" {
			delimiter = '\t'
		}
		w.csvWriter = csv.NewWriter(w.out)
		w.csvWriter.Comma = delimiter
		if err := w.csvWriter.Write(models.Header()); err != nil {
			return fmt.Errorf("写入表头: %w", err)
		}
	default:
		return fmt.Errorf("不支持的格式: %s", w.format)
	}
	w.firstWrite = false
	return nil
}

// writeMetaHeader 在统计输出开头写入版本、编译日期与仓库信息。
func (w *StatsWriter) writeMetaHeader() error {
	switch w.format {
	case "json":
		if _, err := fmt.Fprintf(w.out, "%s\n", version.MetaJSON()); err != nil {
			return err
		}
	default:
		for _, line := range version.HeaderLines() {
			if _, err := io.WriteString(w.out, line+"\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteRow 写入单行统计数据。
func (w *StatsWriter) WriteRow(stats models.ProcessStats) error {
	if err := w.init(); err != nil {
		return err
	}

	switch w.format {
	case "json":
		data, err := json.Marshal(stats)
		if err != nil {
			return fmt.Errorf("序列化 JSON: %w", err)
		}
		if _, err := fmt.Fprintf(w.out, "%s\n", data); err != nil {
			return err
		}
	case "text":
		line := formatTextRow(stats) + "\n"
		if _, err := io.WriteString(w.out, line); err != nil {
			return err
		}
	default:
		gpuIdx := "-1"
		if stats.GPUIndex >= 0 {
			gpuIdx = strconv.Itoa(stats.GPUIndex)
		}
		row := []string{
			stats.Timestamp,
			strconv.Itoa(int(stats.PID)),
			stats.ProcessName,
			fmt.Sprintf("%.2f", stats.CPUPercent),
			fmt.Sprintf("%.3f", stats.MemoryGB),
			gpuIdx,
			fmt.Sprintf("%.2f", stats.GPUMemoryMB),
			fmt.Sprintf("%.1f", stats.GPUUtilization),
			stats.GPUName,
		}
		if err := w.csvWriter.Write(row); err != nil {
			return fmt.Errorf("写入行: %w", err)
		}
		w.csvWriter.Flush()
	}

	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

// WriteRows 批量写入多行统计数据。
func (w *StatsWriter) WriteRows(statsList []models.ProcessStats) error {
	for _, s := range statsList {
		if err := w.WriteRow(s); err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭底层文件句柄（stdout 无需关闭）。
func (w *StatsWriter) Close() error {
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

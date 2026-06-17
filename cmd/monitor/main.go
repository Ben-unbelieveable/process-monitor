/*
进程资源监控主程序

用法:
    monitor -a train.py --tree
    monitor -p 12345 --tree --alert-mem-gb 32 --alert-exit
*/
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/liubo/process-monitor/internal/config"
	"github.com/liubo/process-monitor/internal/monitor"
)

func main() {
	log.SetFlags(0)

	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("[Monitor] 配置错误: %v", err)
	}

	stop := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("\n[Monitor] 收到停止信号，正在退出...")
		close(stop)
	}()

	alerted, err := monitor.Run(cfg, stop)
	if err != nil {
		log.Fatalf("[Monitor] 运行失败: %v", err)
	}
	if alerted {
		os.Exit(2)
	}
}

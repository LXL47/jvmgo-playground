package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LXL47/jvmgo-playground/apps/internal/config"
	"github.com/LXL47/jvmgo-playground/apps/internal/runner"
)

func main() {
	configPath := flag.String("config", "../config/runner.yaml", "配置文件路径")
	flag.Parse()
	cfg, err := config.LoadRunner(*configPath)
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(cfg.WorkRoot, 0o700); err != nil {
		log.Printf("创建工作目录失败: %v", err)
		os.Exit(1)
	}
	server := &http.Server{
		Addr: cfg.Listen, Handler: runner.NewHandler(runner.NewExecutor(cfg), cfg.MaxSourceBytes),
		ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 15 * time.Second,
		IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10,
	}
	log.Printf("Runner 正在监听 %s", cfg.Listen)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

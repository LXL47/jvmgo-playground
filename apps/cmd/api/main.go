package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LXL47/jvmgo-playground/apps/internal/api"
	"github.com/LXL47/jvmgo-playground/apps/internal/config"
)

func main() {
	configPath := flag.String("config", "../config/api.yaml", "配置文件路径")
	flag.Parse()
	cfg, err := config.LoadAPI(*configPath)
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		os.Exit(1)
	}
	server := &http.Server{
		Addr: cfg.Listen, Handler: api.NewHandler(cfg), ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second,
		MaxHeaderBytes: 16 << 10,
	}
	log.Printf("API 正在监听 %s", cfg.Listen)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

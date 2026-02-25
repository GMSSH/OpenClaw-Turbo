package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DemonZack/simplejrpc-go/core"
	"guanxi/eazy-claw/internal/handler/rpc"
)

func main() {
	// Get current working directory
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create temp directory relative to working directory
	tempDir := filepath.Join(workDir, "..", "..", "tmp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		fmt.Printf("Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}

	// Use absolute socket path
	sock, _ := filepath.Abs(filepath.Join(tempDir, "app.sock"))

	// 创建 RPC Server 并注册服务
	server := rpc.NewServer()

	// 启动服务
	if err := server.Start(sock); err != nil {
		core.Container.Log().Error(fmt.Sprintf("启动RPC服务器失败: %v", err))
		return
	}

	core.Container.Log().Info(fmt.Sprintf("EazyClaw 服务启动成功, socket: %s", sock))
}

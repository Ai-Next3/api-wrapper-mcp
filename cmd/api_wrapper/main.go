package main

import (
	"log"
	"os"

	"github.com/gomcpgo/api_wrapper/config"
	"github.com/gomcpgo/api_wrapper/tool"
	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/server"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: api_wrapper <config.yaml>")
	}

	configFile := os.Args[1]
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	apiToolHandler := tool.NewAPIToolHandler(cfg)

	registry := handler.NewHandlerRegistry()
	registry.RegisterToolHandler(apiToolHandler)

	srv := server.New(server.Options{
		Name:     cfg.Server.Name,
		Version:  cfg.Server.Version,
		Registry: registry,
	})

	log.Printf("Starting API Wrapper MCP Server with %d tools...", len(cfg.Tools))
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

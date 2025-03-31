package main

import (
	"log"
	"os"

	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/server"

	"github.com/gomcpgo/api_wrapper/config"
	"github.com/gomcpgo/api_wrapper/tool"
)

func main() {
	// Load configuration from file
	if len(os.Args) < 2 {
		log.Fatal("Usage: api_wrapper <config.yaml>")
	}

	configFile := os.Args[1]
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create API wrapper handler
	apiToolHandler := tool.NewAPIToolHandler(cfg)

	// Register the handler
	registry := handler.NewHandlerRegistry()
	registry.RegisterToolHandler(apiToolHandler)

	// Create and start server
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

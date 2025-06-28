package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/gomcpgo/api_wrapper/config"
	"github.com/gomcpgo/api_wrapper/tool"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	stdioMode := flag.Bool("stdio", false, "Run in stdio mode for local development")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("Usage: api_wrapper [--stdio] <config.yaml>")
	}
	configFile := flag.Arg(0)

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	s := server.NewMCPServer(
		cfg.Server.Name,
		cfg.Server.Version,
		server.WithInstructions(cfg.Server.Description),
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	apiToolHandler := tool.NewAPIToolHandler(cfg)

	tools, err := apiToolHandler.ListTools(context.Background())
	if err != nil {
		log.Fatalf("Failed to list tools from handler: %v", err)
	}

	for _, t := range tools {
		s.AddTool(t, apiToolHandler.CallTool)
	}

	if *stdioMode {
		log.Println("Starting MCP server in stdio mode...")
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("Server error in stdio mode: %v", err)
		}
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8081"
		}
		addr := ":" + port

		log.Printf("Starting StreamableHTTP MCP server on http://localhost%s", addr)
		httpServer := server.NewStreamableHTTPServer(s)

		if err := httpServer.Start(addr); err != nil {
			log.Fatalf("Server error in HTTP mode: %v", err)
		}
	}
}

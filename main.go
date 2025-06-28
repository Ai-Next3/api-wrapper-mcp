package main

import (
	"context"
	"flag"
	"log"
	"os"
	"runtime/debug"

	"github.com/gomcpgo/api_wrapper/config"
	"github.com/gomcpgo/api_wrapper/tool"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Отлов паники
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("!!!!!!!! PANIC DETECTED !!!!!!!!!!\n%v\n%s", r, debug.Stack())
		}
	}()

	log.Println("--- 1. Starting main function ---")

	stdioMode := flag.Bool("stdio", false, "Run in stdio mode for local development")
	flag.Parse()

	log.Println("--- 2. Parsed flags ---")

	if len(flag.Args()) < 1 {
		log.Fatal("Usage: api_wrapper [--stdio] <config.yaml>")
	}
	configFile := flag.Arg(0)
	log.Printf("--- 3. Config file path: %s ---", configFile)

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println("--- 4. Config loaded successfully ---")

	hooks := &server.Hooks{}
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		log.Printf("[HOOK] Received request: Method=%s", method)
	})
	log.Println("--- 5. Hooks created ---")

	s := server.NewMCPServer(
		cfg.Server.Name,
		cfg.Server.Version,
		server.WithInstructions(cfg.Server.Description),
		server.WithToolCapabilities(true),
		server.WithHooks(hooks),
	)
	log.Println("--- 6. MCP Server created ---")

	apiToolHandler := tool.NewAPIToolHandler(cfg)
	log.Println("--- 7. API tool handler created ---")

	tools, err := apiToolHandler.ListTools(context.Background())
	if err != nil {
		log.Fatalf("Failed to list tools from handler: %v", err)
	}
	log.Println("--- 8. Listed tools from handler ---")

	for _, t := range tools {
		s.AddTool(t, apiToolHandler.CallTool)
	}
	log.Println("--- 9. Added tools to server ---")

	if *stdioMode {
		log.Println("--- 10a. Starting in stdio mode... ---")
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("Server error in stdio mode: %v", err)
		}
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8081"
		}
		addr := ":" + port
		log.Printf("--- 10b. Starting in HTTP mode on address %s ---", addr)

		httpServer := server.NewStreamableHTTPServer(s)
		log.Println("--- 11. Streamable HTTP Server created ---")

		if err := httpServer.Start(addr); err != nil {
			log.Fatalf("Server error in HTTP mode: %v", err)
		}
	}
}

package main

import (
	"log"
	"os"

	"github.com/gomcpgo/api_wrapper/config"
	"github.com/gomcpgo/api_wrapper/tool"
	"github.com/mark3labs/mcp-go/server"
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

	// 1. Создаем MCP сервер
	s := server.NewMCPServer(
		cfg.Server.Name,
		cfg.Server.Version,
		server.WithInstructions(cfg.Server.Description),
		server.WithToolCapabilities(true),
	)

	// 2. Создаем наш кастомный обработчик инструментов
	apiToolHandler := tool.NewAPIToolHandler(cfg)

	// 3. Добавляем инструменты на сервер
	tools, err := apiToolHandler.ListTools(nil)
	if err != nil {
		log.Fatalf("Failed to list tools from handler: %v", err)
	}

	for _, t := range tools {
		s.AddTool(t, apiToolHandler.CallTool)
	}

	// 4. Запускаем HTTP сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := ":" + port

	log.Printf("Starting StreamableHTTP MCP server on http://localhost%s", addr)
	httpServer := server.NewStreamableHTTPServer(s)
	if err := httpServer.Start(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

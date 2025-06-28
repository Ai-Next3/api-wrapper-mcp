package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gomcpgo/mcp/pkg/protocol"

	"github.com/gomcpgo/api_wrapper/config"
)

// APIToolHandler implements the ToolHandler interface
type APIToolHandler struct {
	cfg         *config.Config
	toolSchemas map[string]json.RawMessage
}

// NewAPIToolHandler creates a new API tool handler
func NewAPIToolHandler(cfg *config.Config) *APIToolHandler {
	handler := &APIToolHandler{
		cfg:         cfg,
		toolSchemas: make(map[string]json.RawMessage),
	}

	// Generate tool schemas
	for _, tool := range cfg.Tools {
		schema := handler.generateSchema(tool)
		handler.toolSchemas[tool.Name] = schema
	}

	return handler
}

// ListTools returns the list of available tools
func (h *APIToolHandler) ListTools(ctx context.Context) (*protocol.ListToolsResponse, error) {
	tools := make([]protocol.Tool, 0, len(h.cfg.Tools))

	for _, toolCfg := range h.cfg.Tools {
		tool := protocol.Tool{
			Name:        toolCfg.Name,
			Description: toolCfg.Description,
			InputSchema: h.toolSchemas[toolCfg.Name],
		}
		tools = append(tools, tool)
	}

	return &protocol.ListToolsResponse{
		Tools: tools,
	}, nil
}

// CallTool executes an API call
func (h *APIToolHandler) CallTool(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResponse, error) {
	// Find the tool configuration
	var toolCfg *config.ToolConfig
	for i := range h.cfg.Tools {
		if h.cfg.Tools[i].Name == req.Name {
			toolCfg = &h.cfg.Tools[i]
			break
		}
	}

	if toolCfg == nil {
		return &protocol.CallToolResponse{
			IsError: true,
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Tool not found: %s", req.Name),
				},
			},
		}, nil
	}

	// Execute the API call
	result, err := h.executeAPICall(ctx, toolCfg, req.Arguments)
	if err != nil {
		return &protocol.CallToolResponse{
			IsError: true,
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("API call failed: %v", err),
				},
			},
		}, nil
	}

	// Return the result
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: result,
			},
		},
	}, nil
}

// generateSchema generates a JSON schema for a tool
func (h *APIToolHandler) generateSchema(tool config.ToolConfig) json.RawMessage {
	schema := struct {
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties"`
		Required   []string               `json:"required,omitempty"`
	}{
		Type:       "object",
		Properties: make(map[string]interface{}),
		Required:   []string{},
	}

	for name, param := range tool.Parameters {
		propSchema := map[string]interface{}{
			"type":        param.Type,
			"description": param.Description,
		}

		if param.Default != nil {
			propSchema["default"] = param.Default
		}

		if len(param.Enum) > 0 {
			propSchema["enum"] = param.Enum
		}

		schema.Properties[name] = propSchema

		if param.Required {
			schema.Required = append(schema.Required, name)
		}
	}

	schemaJSON, _ := json.Marshal(schema)
	return json.RawMessage(schemaJSON)
}

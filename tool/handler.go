package tool

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/gomcpgo/api_wrapper/config"
)

// APIToolHandler реализует логику для работы с инструментами API
type APIToolHandler struct {
	cfg *config.Config
}

// NewAPIToolHandler создает новый обработчик
func NewAPIToolHandler(cfg *config.Config) *APIToolHandler {
	return &APIToolHandler{
		cfg: cfg,
	}
}

// ListTools возвращает список доступных инструментов в формате новой библиотеки
func (h *APIToolHandler) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	var tools []mcp.Tool
	for _, toolCfg := range h.cfg.Tools {
		var toolOptions []mcp.ToolOption
		toolOptions = append(toolOptions, mcp.WithDescription(toolCfg.Description))

		for name, param := range toolCfg.Parameters {
			var paramOptions []mcp.PropertyOption
			paramOptions = append(paramOptions, mcp.Description(param.Description))
			if param.Required {
				paramOptions = append(paramOptions, mcp.Required())
			}
			if param.Default != nil {
				switch v := param.Default.(type) {
				case string:
					paramOptions = append(paramOptions, mcp.DefaultString(v))
				case int, int8, int16, int32, int64, float32, float64:
					paramOptions = append(paramOptions, mcp.DefaultNumber(v.(float64)))
				case bool:
					paramOptions = append(paramOptions, mcp.DefaultBool(v))
				}
			}
			// Добавляем параметр в инструмент в зависимости от типа
			switch param.Type {
			case "string":
				toolOptions = append(toolOptions, mcp.WithString(name, paramOptions...))
			case "number":
				toolOptions = append(toolOptions, mcp.WithNumber(name, paramOptions...))
			case "boolean":
				toolOptions = append(toolOptions, mcp.WithBoolean(name, paramOptions...))
			}
		}

		tool := mcp.NewTool(toolCfg.Name, toolOptions...)
		tools = append(tools, tool)
	}
	return tools, nil
}

// CallTool вызывает инструмент API
func (h *APIToolHandler) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var toolCfg *config.ToolConfig
	for i := range h.cfg.Tools {
		if h.cfg.Tools[i].Name == req.Params.Name {
			toolCfg = &h.cfg.Tools[i]
			break
		}
	}

	if toolCfg == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tool not found: %s", req.Params.Name)), nil
	}

	args := req.GetArguments()
	result, err := h.executeAPICall(ctx, toolCfg, args)
	if err != nil {
		return mcp.NewToolResultErrorFromErr(fmt.Sprintf("API call failed for tool %s", toolCfg.Name), err), nil
	}

	return mcp.NewToolResultText(result), nil
}

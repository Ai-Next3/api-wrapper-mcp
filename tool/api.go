package tool

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/gomcpgo/api_wrapper/config"
)

// executeAPICall выполняет API-вызов на основе конфигурации инструмента и аргументов
func (h *APIToolHandler) executeAPICall(ctx context.Context, toolCfg *config.ToolConfig, args map[string]interface{}) (string, error) {
	// Применяем значения по умолчанию для отсутствующих аргументов
	for name, param := range toolCfg.Parameters {
		if _, exists := args[name]; !exists && param.Default != nil {
			args[name] = param.Default
		}
	}

	// Создаем новый HTTP-клиент с таймаутом
	client := &http.Client{
		Timeout: time.Duration(toolCfg.Timeout) * time.Second,
	}

	var req *http.Request
	var err error

	// Получаем API-токен из переменных окружения
	apiToken := os.Getenv(h.cfg.Auth.TokenEnvVar)

	switch toolCfg.Method {
	case "GET":
		reqURL, err := url.Parse(toolCfg.Endpoint)
		if err != nil {
			return "", fmt.Errorf("invalid endpoint URL: %w", err)
		}

		query := reqURL.Query()
		for key, tmplVal := range toolCfg.QueryParams {
			val, err := h.processTemplate(tmplVal, args)
			if err != nil {
				return "", fmt.Errorf("failed to process query parameter '%s': %w", key, err)
			}
			query.Add(key, val)
		}
		reqURL.RawQuery = query.Encode()

		req, err = http.NewRequestWithContext(ctx, "GET", reqURL.String(), nil)
		if err != nil {
			return "", fmt.Errorf("failed to create GET request: %w", err)
		}

	case "POST":
		jsonBody, err := h.processTemplate(toolCfg.Template, args)
		if err != nil {
			return "", fmt.Errorf("failed to process request template: %w", err)
		}

		req, err = http.NewRequestWithContext(ctx, "POST", toolCfg.Endpoint, bytes.NewBufferString(jsonBody))
		if err != nil {
			return "", fmt.Errorf("failed to create POST request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

	default:
		return "", fmt.Errorf("unsupported HTTP method: %s", toolCfg.Method)
	}

	// Устанавливаем заголовок авторизации, если токен предоставлен
	if apiToken != "" {
		req.Header.Set("Authorization", apiToken)
	}

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API returned error status %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

// processTemplate обрабатывает строку шаблона с заданными аргументами
func (h *APIToolHandler) processTemplate(tmplStr string, args map[string]interface{}) (string, error) {
	// Простой обработчик для замены {{variable}}
	tmpl, err := template.New("").Delims("{{", "}}").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}

	// Специальная обработка для синтаксиса {{env:VARIABLE}}
	processedArgs := make(map[string]interface{})
	for k, v := range args {
		processedArgs[k] = v
	}

	// Заменяем переменные окружения в шаблоне
	for k, v := range processedArgs {
		if strVal, ok := v.(string); ok {
			if strings.HasPrefix(strVal, "{{env:") && strings.HasSuffix(strVal, "}}") {
				envVar := strings.TrimPrefix(strings.TrimSuffix(strVal, "}}"), "{{env:")
				envVal := os.Getenv(envVar)
				processedArgs[k] = envVal
			}
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, processedArgs); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return buf.String(), nil
}

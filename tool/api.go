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

// executeAPICall executes an API call based on the tool configuration and arguments
func (h *APIToolHandler) executeAPICall(ctx context.Context, toolCfg *config.ToolConfig, args map[string]interface{}) (string, error) {
	// Apply default values for missing arguments
	for name, param := range toolCfg.Parameters {
		if _, exists := args[name]; !exists && param.Default != nil {
			args[name] = param.Default
		}
	}

	// Create a new HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(toolCfg.Timeout) * time.Second,
	}

	var req *http.Request
	var err error

	// Get API token from environment
	apiToken := os.Getenv(h.cfg.Auth.TokenEnvVar)

	switch toolCfg.Method {
	case "GET":
		// Build query parameters for GET request
		reqURL, err := url.Parse(toolCfg.Endpoint)
		if err != nil {
			return "", fmt.Errorf("invalid endpoint URL: %w", err)
		}

		query := reqURL.Query()
		for key, tmplVal := range toolCfg.QueryParams {
			// Process template values in query params
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
		// Process the JSON template for POST request
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

	// Set authorization header if API token is provided
	if apiToken != "" {
		// req.Header.Set("Authorization", "Bearer "+apiToken)
		req.Header.Set("Authorization", apiToken)
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error status code
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API returned error status %d: %s", resp.StatusCode, string(body))
	}

	// For successful responses, return the body as-is
	return string(body), nil
}

// processTemplate processes a template string with the given arguments
func (h *APIToolHandler) processTemplate(tmplStr string, args map[string]interface{}) (string, error) {
	// Simple template processor for {{variable}} replacement
	tmpl, err := template.New("").Delims("{{", "}}").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}

	// Special handling for {{env:VARIABLE}} syntax
	// Create a copy of args to avoid modifying the original
	processedArgs := make(map[string]interface{})
	for k, v := range args {
		processedArgs[k] = v
	}

	// Process env vars in the template
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

package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config represents the API wrapper configuration
type Config struct {
	Server struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Version     string `yaml:"version"`
	} `yaml:"server"`

	Auth struct {
		TokenEnvVar string `yaml:"token_env_var"`
	} `yaml:"auth"`

	Tools []ToolConfig `yaml:"tools"`
}

// ToolConfig represents a single tool configuration
type ToolConfig struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Endpoint    string                 `yaml:"endpoint"`
	Method      string                 `yaml:"method"`
	Timeout     int                    `yaml:"timeout"`
	Template    string                 `yaml:"template"`
	QueryParams map[string]string      `yaml:"query_params,omitempty"`
	Parameters  map[string]ParamConfig `yaml:"parameters"`
}

// ParamConfig represents a parameter configuration
type ParamConfig struct {
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	Required    bool        `yaml:"required"`
	Default     interface{} `yaml:"default,omitempty"`
	Enum        []string    `yaml:"enum,omitempty"`
}

// LoadConfig loads the API wrapper configuration from a file
func LoadConfig(filePath string) (*Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.Server.Name == "" {
		cfg.Server.Name = "API Wrapper MCP"
	}
	if cfg.Server.Version == "" {
		cfg.Server.Version = "1.0.0"
	}

	// Validate configuration
	for i, tool := range cfg.Tools {
		if tool.Name == "" {
			return nil, fmt.Errorf("tool at index %d has no name", i)
		}
		if tool.Endpoint == "" {
			return nil, fmt.Errorf("tool '%s' has no endpoint", tool.Name)
		}
		if tool.Method != "GET" && tool.Method != "POST" {
			return nil, fmt.Errorf("tool '%s' has unsupported method: %s (must be GET or POST)", tool.Name, tool.Method)
		}
		if tool.Timeout <= 0 {
			cfg.Tools[i].Timeout = 30 // Default timeout of 30 seconds
		}
	}

	return &cfg, nil
}

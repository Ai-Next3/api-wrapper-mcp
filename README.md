# API Wrapper MCP Server

This is a generic API wrapper server for the Model Context Protocol (MCP). It allows you to easily wrap REST APIs as MCP tools that can be accessed by Claude and other MCP clients.

## Features

- Easy YAML configuration for multiple API endpoints
- Support for both GET and POST requests
- Parameter validation and default values
- Authentication via environment variables
- Custom timeouts for API calls

## Usage

1. Create a YAML configuration file defining your API endpoints (see `example-config.yaml`)
2. Set any required API tokens as environment variables
3. Run the server with your config file:

```bash
# Build the server
go build -o api_wrapper

# Run with your config
./api_wrapper my-apis.yaml
```

## Configuration Format

The configuration file uses YAML format with the following structure:

```yaml
# Server info
server:
  name: "API Gateway MCP"
  description: "Generic API gateway that wraps REST APIs as MCP tools"
  version: "1.0.0"
  
# Authentication
auth:
  token_env_var: "API_GATEWAY_TOKEN"  # Environment variable for the API token

# Tool definitions
tools:
  - name: "tool-name"
    description: "Tool description"
    endpoint: "https://api.example.com/endpoint"
    method: "POST"  # or "GET"
    timeout: 30  # in seconds
    template: |
      {
        "param1": "{{variable1}}",
        "param2": {{variable2}}
      }
    # For GET requests, use query_params
    query_params:
      param1: "{{variable1}}"
      param2: "{{variable2}}"
    parameters:
      variable1:
        type: "string"
        description: "Description of variable1"
        required: true
      variable2:
        type: "number"
        description: "Description of variable2"
        default: 10
```

## Claude Desktop Integration

To use with Claude Desktop, add the following to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "api-wrapper": {
      "command": "path/to/api_wrapper",
      "args": ["path/to/your-config.yaml"],
      "env": {
        "API_GATEWAY_TOKEN": "your-api-token"
      }
    }
  }
}
```

## Examples

Check out `example-config.yaml` for sample API configurations.

## Environment Variables

- Set the main authentication token using the environment variable specified in the `auth.token_env_var` field.
- You can also reference other environment variables in your templates using `{{env:VARIABLE_NAME}}` syntax.

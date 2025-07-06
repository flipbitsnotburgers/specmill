# Specmill Usage Guide

## Basic Usage

```bash
./specmill-server -spec path/to/openapi.yaml
```

## How It Works

1. **Reads OpenAPI spec** from the YAML file
2. **Extracts server URL** from the `servers` section
3. **Generates MCP tools** for each operation with an `operationId`
4. **Proxies requests** - when an MCP client calls a tool, Specmill makes the HTTP request to the actual API

## OpenAPI Requirements

Your OpenAPI spec needs:
- Valid `servers` section with URLs
- Operations with unique `operationId` values
- Proper parameter definitions

Example:
```yaml
openapi: 3.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /users/{id}:
    get:
      operationId: getUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
```

## Common Issues

### Relative URLs
If your OpenAPI spec has relative URLs like `/api/v3`, the HTTP requests won't work. Solutions:
- Edit the spec to use absolute URLs
- Use a spec hosted on the actual API server

### Missing operationId
Tools are only generated for operations that have an `operationId`. Add one to each operation you want to expose.

### Authentication
Currently no built-in auth support. You can:
- Use publicly accessible APIs
- Add auth tokens via environment variables (future feature)

## Testing

Test that your server works:
```bash
# List available tools
echo '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":1}' | ./specmill-server -spec your-api.yaml

# Should return a list of tools based on your OpenAPI operations
```
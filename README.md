# Specmill

An MCP (Model Context Protocol) proxy server that exposes any OpenAPI-compliant REST API as MCP tools for LLMs.

## What It Does

Specmill acts as a **proxy between LLMs and REST APIs**:

1. **Reads an OpenAPI specification** (YAML format) describing a REST API
2. **Generates MCP tools** from each API operation in the spec
3. **When an LLM calls a tool**, Specmill makes the actual HTTP request to the API
4. **Returns the API response** back to the LLM

This allows any MCP-compatible LLM to interact with REST APIs that have an OpenAPI specification without writing custom MCP server code.

## Features

- **API Proxy** - Makes real HTTP requests to the API endpoints
- **Automatic Tool Generation** - Creates MCP tools from OpenAPI operations
- **Parameter Mapping** - Converts function arguments to HTTP parameters
- **Schema Validation** - Uses OpenAPI schemas for parameter types
- **Multiple APIs** - Run multiple instances for different APIs

## Quick Start

### Build the project

```bash
make build
```

### Run with an OpenAPI spec

```bash
./specmill-server -spec examples/petstore.yaml
```

### Use with Claude or other MCP clients

```bash
# Initialize the server
echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2025-06-18"},"id":1}' | ./specmill-server -spec examples/petstore.yaml

# List available tools
echo '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":2}' | ./specmill-server -spec examples/petstore.yaml

# Call a tool
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"getPetById","arguments":{"petId":1}},"id":3}' | ./specmill-server -spec examples/petstore.yaml
```

## Project Structure

- `parser/` - OpenAPI specification parser
  - `openapi.go` - YAML parsing and schema definitions
- `generator/` - MCP server generator
  - `types.go` - MCP protocol type definitions
  - `mcp.go` - OpenAPI to MCP conversion logic
- `server/` - MCP server implementation
  - `server.go` - JSON-RPC server and request handling
- `examples/` - Example OpenAPI specifications
  - `petstore.yaml` - Standard Petstore API for testing
- `main.go` - CLI entry point

## How It Works

When an LLM calls a tool (e.g., `getPetById`), Specmill:

1. **Receives the MCP tool call** with arguments from the LLM
2. **Maps it to the OpenAPI operation** (e.g., `GET /pet/{petId}`)
3. **Constructs the HTTP request**:
   - Uses the base URL from the OpenAPI `servers` section
   - Substitutes path parameters
   - Adds query parameters
   - Sets request body (for POST/PUT)
4. **Makes the actual HTTP request** to the API server
5. **Returns the response** to the LLM

### Example Flow

```
LLM: "Get pet with ID 123"
  ↓
MCP: tools/call { name: "getPetById", arguments: { petId: 123 } }
  ↓
Specmill: GET [servers.url]/pet/123
  ↓
API: { "id": 123, "name": "Fluffy", "status": "available" }
  ↓
LLM: "The pet with ID 123 is named Fluffy and is available"
```

### OpenAPI to MCP Mapping

| OpenAPI | MCP | HTTP |
|---------|-----|------|
| Operation | Tool | HTTP Method |
| OperationId | Tool Name | - |
| Path + Parameters | - | URL Construction |
| RequestBody | Tool Arguments | Request Body |
| Responses | Tool Response | Response Body |

## Development

### Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile)

### Testing

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose

# Run with coverage
make test-coverage
```

### Available Make Targets

- `build` - Build the server binary
- `test` - Run unit tests
- `test-verbose` - Run tests with detailed output
- `test-coverage` - Run tests with coverage report
- `clean` - Remove build artifacts
- `fmt` - Format Go code
- `lint` - Run linter (requires golangci-lint)
- `help` - Show available targets

## Use Cases

Specmill is perfect for:

1. **Testing APIs** - Let LLMs interact with your development APIs
2. **API Integration** - Give LLMs access to third-party APIs (GitHub, Stripe, etc.)
3. **Internal Tools** - Expose internal REST APIs to LLMs without writing code
4. **API Exploration** - Use LLMs to explore and understand new APIs
5. **Automation** - Let LLMs perform API operations based on natural language

## Example: Petstore API

When running with the Petstore example, Specmill:

1. **Connects to** the live Petstore API at `https://petstore3.swagger.io`
2. **Generates MCP tools** like:
   - `addPet` - Creates a new pet (POST /pet)
   - `getPetById` - Fetches a pet (GET /pet/{petId})
   - `updatePet` - Updates a pet (PUT /pet)
   - `deletePet` - Deletes a pet (DELETE /pet/{petId})
3. **Makes real HTTP requests** when the LLM uses these tools
4. **Returns actual API responses** to the LLM

Note: The Petstore spec uses a relative URL (`/api/v3`), so you'll need an OpenAPI spec with absolute URLs for real API calls.

## VSCode Integration

Specmill can be integrated with VSCode as an MCP server to use OpenAPI-based tools with Claude.

### Setup

1. **Build the server**
   ```bash
   make build
   ```

2. **Create MCP configuration**
   
   Create `.vscode/mcp.json` in your project root:
   ```json
   {
     "mcpServers": {
       "petstore": {
         "command": "/absolute/path/to/specmill-server",
         "args": ["-spec", "/absolute/path/to/examples/petstore.yaml"],
         "env": {},
         "disabled": false
       }
     }
   }
   ```

### Documentation

- **[VSCode Setup](docs/VSCODE_SETUP.md)** - Quick VSCode configuration guide
- **[Usage Guide](docs/USAGE.md)** - Detailed usage instructions and troubleshooting
- **[Examples](examples/)** - Example configurations and OpenAPI specs

## TODO

- [ ] Support for authentication schemes (API keys, OAuth, etc.)
- [ ] Handle non-JSON request/response content types
- [ ] Add response parsing and formatting
- [ ] Support for OpenAPI 3.1 features
- [ ] Configuration for base URLs and defaults
- [ ] Better error messages and validation

## Reference

- [Model Context Protocol](https://modelcontextprotocol.io/)
- [OpenAPI Specification](https://www.openapis.org/)
- [JSON-RPC 2.0](https://www.jsonrpc.org/specification)
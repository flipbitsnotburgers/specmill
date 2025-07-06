# VSCode MCP Integration

Quick guide to set up Specmill as an MCP server in VSCode.

## Prerequisites

- Specmill built (`make build`)
- VSCode with MCP extension
- OpenAPI spec files (YAML)

## Setup

1. **Create `.vscode/mcp.json`** in your workspace:

```json
{
  "mcpServers": {
    "petstore": {
      "command": "/absolute/path/to/specmill-server",
      "args": ["-spec", "/absolute/path/to/openapi.yaml"],
      "env": {},
      "disabled": false
    }
  }
}
```

2. **Restart VSCode** for changes to take effect

## Configuration Fields

- `command`: Absolute path to specmill-server binary
- `args`: Must include `-spec` and path to OpenAPI YAML
- `env`: Optional environment variables
- `disabled`: Set to `true` to disable temporarily

## Multiple APIs Example

```json
{
  "mcpServers": {
    "petstore": {
      "command": "/home/user/specmill/specmill-server",
      "args": ["-spec", "/home/user/apis/petstore.yaml"],
      "disabled": false
    },
    "github": {
      "command": "/home/user/specmill/specmill-server",
      "args": ["-spec", "/home/user/apis/github.yaml"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      },
      "disabled": false
    }
  }
}
```

## Troubleshooting

- **Not showing up**: Check paths are absolute, not relative
- **Not working**: Test manually: `echo '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":1}' | /path/to/specmill-server -spec /path/to/spec.yaml`
- **Check logs**: View > Output > MCP

## Examples

See `examples/vscode/` for complete configuration examples.
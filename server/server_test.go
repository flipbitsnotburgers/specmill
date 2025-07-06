package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"specmill/generator"
)

func TestMCPServer(t *testing.T) {
	// Create server with Petstore spec
	srv, err := NewMCPServer("../examples/petstore.yaml")
	if err != nil {
		t.Fatalf("Failed to create MCP server: %v", err)
	}

	tests := []struct {
		name     string
		request  string
		validate func(t *testing.T, response *generator.MCPResponse)
	}{
		{
			name: "Initialize",
			request: `{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}},"id":1}`,
			validate: func(t *testing.T, response *generator.MCPResponse) {
				if response.Error != nil {
					t.Fatalf("Expected no error, got: %v", response.Error)
				}

				var result generator.InitializeResult
				if err := json.Unmarshal(response.Result, &result); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}

				if result.ProtocolVersion != "2025-06-18" {
					t.Errorf("Expected protocol version 2025-06-18, got: %s", result.ProtocolVersion)
				}

				if result.ServerInfo.Name != "specmill" {
					t.Errorf("Expected server name 'specmill', got: %s", result.ServerInfo.Name)
				}
			},
		},
		{
			name: "List Tools",
			request: `{"jsonrpc":"2.0","method":"tools/list","params":{},"id":2}`,
			validate: func(t *testing.T, response *generator.MCPResponse) {
				if response.Error != nil {
					t.Fatalf("Expected no error, got: %v", response.Error)
				}

				var result generator.ListToolsResult
				if err := json.Unmarshal(response.Result, &result); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}

				if len(result.Tools) == 0 {
					t.Error("Expected at least one tool, got none")
				}

				// Check for some expected Petstore tools
				toolNames := make(map[string]bool)
				for _, tool := range result.Tools {
					toolNames[tool.Name] = true
				}

				expectedTools := []string{"getPetById", "addPet", "updatePet", "deletePet"}
				for _, expected := range expectedTools {
					if !toolNames[expected] {
						t.Errorf("Expected tool '%s' not found", expected)
					}
				}

				// Validate a specific tool structure
				for _, tool := range result.Tools {
					if tool.Name == "getPetById" {
						// Check that it has required parameters
						var schema map[string]interface{}
						if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
							t.Errorf("Failed to unmarshal getPetById schema: %v", err)
							continue
						}

						props, ok := schema["properties"].(map[string]interface{})
						if !ok {
							t.Error("getPetById schema missing properties")
							continue
						}

						if _, ok := props["petId"]; !ok {
							t.Error("getPetById schema missing petId parameter")
						}

						required, ok := schema["required"].([]interface{})
						if !ok || len(required) == 0 {
							t.Error("getPetById should have required parameters")
						}
					}
				}
			},
		},
		{
			name: "Invalid Method",
			request: `{"jsonrpc":"2.0","method":"invalid/method","params":{},"id":3}`,
			validate: func(t *testing.T, response *generator.MCPResponse) {
				if response.Error == nil {
					t.Fatal("Expected error for invalid method")
				}

				if response.Error.Code != -32601 {
					t.Errorf("Expected error code -32601, got: %d", response.Error.Code)
				}
			},
		},
		{
			name: "Call Tool with Invalid Params",
			request: `{"jsonrpc":"2.0","method":"tools/call","params":"invalid","id":4}`,
			validate: func(t *testing.T, response *generator.MCPResponse) {
				if response.Error == nil {
					t.Fatal("Expected error for invalid params")
				}

				if response.Error.Code != -32602 {
					t.Errorf("Expected error code -32602, got: %d", response.Error.Code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse request
			var request generator.MCPRequest
			if err := json.Unmarshal([]byte(tt.request), &request); err != nil {
				t.Fatalf("Failed to parse request: %v", err)
			}

			// Handle request
			response := srv.handleRequest(&request)

			// Validate response
			tt.validate(t, response)
		})
	}
}

func TestMCPServerInvalidSpec(t *testing.T) {
	_, err := NewMCPServer("nonexistent.yaml")
	if err == nil {
		t.Fatal("Expected error for non-existent spec file")
	}
}

func TestMCPServerStdio(t *testing.T) {
	srv, err := NewMCPServer("../examples/petstore.yaml")
	if err != nil {
		t.Fatalf("Failed to create MCP server: %v", err)
	}

	// Test writing response
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	response := &generator.MCPResponse{
		Jsonrpc: "2.0",
		Result:  json.RawMessage(`{"test":"value"}`),
		ID:      1,
	}

	if err := srv.writeResponse(writer, response); err != nil {
		t.Fatalf("Failed to write response: %v", err)
	}

	output := buf.String()
	if !strings.HasSuffix(output, "\n") {
		t.Error("Response should end with newline")
	}

	// Check that it's valid JSON
	var parsed generator.MCPResponse
	jsonOutput := strings.TrimSuffix(output, "\n")
	if err := json.Unmarshal([]byte(jsonOutput), &parsed); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}
}
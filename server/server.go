package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"specmill/generator"
	"specmill/parser"
)

type MCPServer struct {
	generator *generator.MCPGenerator
	spec      *parser.OpenAPISpec
}

func NewMCPServer(specPath string) (*MCPServer, error) {
	spec, err := parser.ParseOpenAPISpec(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	gen := generator.NewMCPGenerator(spec)
	if err := gen.GenerateTools(); err != nil {
		return nil, fmt.Errorf("failed to generate tools: %w", err)
	}

	return &MCPServer{
		generator: gen,
		spec:      spec,
	}, nil
}

func (s *MCPServer) Start() error {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var request generator.MCPRequest
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			s.writeError(writer, nil, -32700, "Parse error")
			continue
		}

		response := s.handleRequest(&request)
		if err := s.writeResponse(writer, response); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

func (s *MCPServer) handleRequest(request *generator.MCPRequest) *generator.MCPResponse {
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "tools/list":
		return s.handleListTools(request)
	case "tools/call":
		return s.handleCallTool(request)
	default:
		return &generator.MCPResponse{
			Jsonrpc: "2.0",
			Error: &generator.MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
			ID: request.ID,
		}
	}
}

func (s *MCPServer) handleInitialize(request *generator.MCPRequest) *generator.MCPResponse {
	result := generator.InitializeResult{
		ProtocolVersion: "2025-06-18",
		Capabilities: generator.Capabilities{
			Tools: map[string]any{},
		},
		ServerInfo: generator.ServerInfo{
			Name:    "specmill",
			Version: "1.0.0",
		},
	}

	resultBytes, _ := json.Marshal(result)
	return &generator.MCPResponse{
		Jsonrpc: "2.0",
		Result:  json.RawMessage(resultBytes),
		ID:      request.ID,
	}
}

func (s *MCPServer) handleListTools(request *generator.MCPRequest) *generator.MCPResponse {
	result := generator.ListToolsResult{
		Tools: s.generator.GetTools(),
	}

	resultBytes, _ := json.Marshal(result)
	return &generator.MCPResponse{
		Jsonrpc: "2.0",
		Result:  json.RawMessage(resultBytes),
		ID:      request.ID,
	}
}

func (s *MCPServer) handleCallTool(request *generator.MCPRequest) *generator.MCPResponse {
	var params generator.CallToolParams
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return &generator.MCPResponse{
			Jsonrpc: "2.0",
			Error: &generator.MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
			ID: request.ID,
		}
	}

	result, err := s.generator.ExecuteTool(params.Name, params.Arguments)
	if err != nil {
		return &generator.MCPResponse{
			Jsonrpc: "2.0",
			Error: &generator.MCPError{
				Code:    -32603,
				Message: err.Error(),
			},
			ID: request.ID,
		}
	}

	resultBytes, _ := json.Marshal(result)
	return &generator.MCPResponse{
		Jsonrpc: "2.0",
		Result:  json.RawMessage(resultBytes),
		ID:      request.ID,
	}
}

func (s *MCPServer) writeResponse(w *bufio.Writer, response *generator.MCPResponse) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}

	return w.Flush()
}

func (s *MCPServer) writeError(w *bufio.Writer, id interface{}, code int, message string) {
	response := &generator.MCPResponse{
		Jsonrpc: "2.0",
		Error: &generator.MCPError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
	_ = s.writeResponse(w, response)
}
package generator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"specmill/parser"
)

type MCPGenerator struct {
	spec      *parser.OpenAPISpec
	tools     []MCPTool
	baseURL   string
	client    *http.Client
}

func NewMCPGenerator(spec *parser.OpenAPISpec) *MCPGenerator {
	baseURL := ""
	if len(spec.Servers) > 0 {
		baseURL = spec.Servers[0].URL
	}

	return &MCPGenerator{
		spec:    spec,
		tools:   []MCPTool{},
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (g *MCPGenerator) GenerateTools() error {
	for path, pathItem := range g.spec.Paths {
		operations := pathItem.GetOperations()
		for method, operation := range operations {
			if operation.OperationID == "" {
				continue
			}

			tool := MCPTool{
				Name:        operation.OperationID,
				Description: g.generateDescription(operation, method, path),
				InputSchema: g.generateInputSchema(operation),
			}

			g.tools = append(g.tools, tool)
		}
	}

	return nil
}

func (g *MCPGenerator) generateDescription(op *parser.Operation, method, path string) string {
	desc := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
	if op.Summary != "" {
		desc = op.Summary
	}
	if op.Description != "" {
		desc = fmt.Sprintf("%s\n%s", desc, op.Description)
	}
	return desc
}

func (g *MCPGenerator) generateInputSchema(op *parser.Operation) json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": make(map[string]interface{}),
		"required": []string{},
	}

	properties := schema["properties"].(map[string]interface{})
	required := []string{}

	for _, param := range op.Parameters {
		paramSchema := g.convertSchema(param.Schema)
		if paramSchema == nil {
			paramSchema = map[string]interface{}{"type": "string"}
		}

		paramSchema.(map[string]interface{})["description"] = param.Description
		
		paramName := param.Name
		if param.In != "query" && param.In != "path" {
			paramName = param.In + "_" + param.Name
		}

		properties[paramName] = paramSchema

		if param.Required {
			required = append(required, paramName)
		}
	}

	if op.RequestBody != nil && op.RequestBody.Content != nil {
		for contentType, mediaType := range op.RequestBody.Content {
			if strings.Contains(contentType, "json") && mediaType.Schema != nil {
				properties["body"] = g.convertSchema(mediaType.Schema)
				if op.RequestBody.Required {
					required = append(required, "body")
				}
				break
			}
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	} else {
		delete(schema, "required")
	}

	data, _ := json.Marshal(schema)
	return json.RawMessage(data)
}

func (g *MCPGenerator) convertSchema(schema *parser.Schema) interface{} {
	if schema == nil {
		return nil
	}

	if schema.Ref != "" {
		refParts := strings.Split(schema.Ref, "/")
		if len(refParts) > 0 && refParts[len(refParts)-2] == "schemas" {
			schemaName := refParts[len(refParts)-1]
			if g.spec.Components != nil && g.spec.Components.Schemas != nil {
				if refSchema, ok := g.spec.Components.Schemas[schemaName]; ok {
					return g.convertSchema(refSchema)
				}
			}
		}
		return map[string]interface{}{"type": "object"}
	}

	result := map[string]interface{}{
		"type": schema.Type,
	}

	if schema.Format != "" {
		result["format"] = schema.Format
	}

	if schema.Description != "" {
		result["description"] = schema.Description
	}

	if len(schema.Enum) > 0 {
		result["enum"] = schema.Enum
	}

	if schema.Type == "object" && len(schema.Properties) > 0 {
		properties := make(map[string]interface{})
		for name, propSchema := range schema.Properties {
			properties[name] = g.convertSchema(propSchema)
		}
		result["properties"] = properties

		if len(schema.Required) > 0 {
			result["required"] = schema.Required
		}
	}

	if schema.Type == "array" && schema.Items != nil {
		result["items"] = g.convertSchema(schema.Items)
	}

	return result
}

func (g *MCPGenerator) GetTools() []MCPTool {
	return g.tools
}

func (g *MCPGenerator) ExecuteTool(name string, arguments json.RawMessage) (*CallToolResult, error) {
	var path string
	var method string
	var operation *parser.Operation

	for p, pathItem := range g.spec.Paths {
		operations := pathItem.GetOperations()
		for m, op := range operations {
			if op.OperationID == name {
				path = p
				method = m
				operation = op
				break
			}
		}
		if operation != nil {
			break
		}
	}

	if operation == nil {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	var args map[string]interface{}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	url := g.baseURL + path
	for _, param := range operation.Parameters {
		if param.In == "path" {
			if value, ok := args[param.Name]; ok {
				placeholder := fmt.Sprintf("{%s}", param.Name)
				url = strings.Replace(url, placeholder, fmt.Sprint(value), 1)
				delete(args, param.Name)
			}
		}
	}

	req, err := http.NewRequest(strings.ToUpper(method), url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	for _, param := range operation.Parameters {
		if param.In == "query" {
			if value, ok := args[param.Name]; ok {
				q.Add(param.Name, fmt.Sprint(value))
				delete(args, param.Name)
			}
		}
	}
	req.URL.RawQuery = q.Encode()

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return &CallToolResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("HTTP %d response received", resp.StatusCode),
			},
		},
	}, nil
}
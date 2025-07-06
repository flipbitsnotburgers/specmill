package generator

import (
	"encoding/json"
	"testing"

	"specmill/parser"
)

func TestNewMCPGenerator(t *testing.T) {
	spec := &parser.OpenAPISpec{
		Servers: []parser.Server{
			{URL: "https://api.example.com"},
		},
	}

	gen := NewMCPGenerator(spec)

	if gen.spec != spec {
		t.Error("Spec not correctly assigned")
	}

	if gen.baseURL != "https://api.example.com" {
		t.Errorf("Expected baseURL 'https://api.example.com', got: %s", gen.baseURL)
	}

	if gen.client == nil {
		t.Error("HTTP client should be initialized")
	}
}

func TestGenerateTools(t *testing.T) {
	spec := &parser.OpenAPISpec{
		Paths: map[string]parser.PathItem{
			"/pets": {
				Get: &parser.Operation{
					OperationID: "listPets",
					Summary:     "List all pets",
					Parameters: []parser.Parameter{
						{
							Name:     "limit",
							In:       "query",
							Required: false,
							Schema:   &parser.Schema{Type: "integer"},
						},
					},
				},
				Post: &parser.Operation{
					OperationID: "createPet",
					Summary:     "Create a pet",
					RequestBody: &parser.RequestBody{
						Required: true,
						Content: map[string]parser.MediaType{
							"application/json": {
								Schema: &parser.Schema{
									Type: "object",
									Properties: map[string]*parser.Schema{
										"name": {Type: "string"},
										"tag":  {Type: "string"},
									},
									Required: []string{"name"},
								},
							},
						},
					},
				},
			},
		},
	}

	gen := NewMCPGenerator(spec)
	err := gen.GenerateTools()
	if err != nil {
		t.Fatalf("Failed to generate tools: %v", err)
	}

	tools := gen.GetTools()
	if len(tools) != 2 {
		t.Fatalf("Expected 2 tools, got: %d", len(tools))
	}

	// Check listPets tool
	var listPetsTool *MCPTool
	for i := range tools {
		if tools[i].Name == "listPets" {
			listPetsTool = &tools[i]
			break
		}
	}

	if listPetsTool == nil {
		t.Fatal("listPets tool not found")
	}

	if listPetsTool.Description != "List all pets" {
		t.Errorf("Expected description 'List all pets', got: %s", listPetsTool.Description)
	}

	// Check schema
	var schema map[string]interface{}
	err = json.Unmarshal(listPetsTool.InputSchema, &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["limit"]; !ok {
		t.Error("Schema should have 'limit' property")
	}

	// Check createPet tool
	var createPetTool *MCPTool
	for i := range tools {
		if tools[i].Name == "createPet" {
			createPetTool = &tools[i]
			break
		}
	}

	if createPetTool == nil {
		t.Fatal("createPet tool not found")
	}

	// Check request body in schema
	err = json.Unmarshal(createPetTool.InputSchema, &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal createPet schema: %v", err)
	}

	props, ok = schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	if _, ok := props["body"]; !ok {
		t.Error("Schema should have 'body' property for request body")
	}

	required, ok := schema["required"].([]interface{})
	if !ok || len(required) == 0 {
		t.Error("Schema should have required fields")
	}
}

func TestConvertSchema(t *testing.T) {
	gen := NewMCPGenerator(&parser.OpenAPISpec{})

	tests := []struct {
		name     string
		input    *parser.Schema
		validate func(t *testing.T, result interface{})
	}{
		{
			name: "Simple string schema",
			input: &parser.Schema{
				Type:        "string",
				Description: "A test string",
			},
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("Result should be a map")
				}
				if m["type"] != "string" {
					t.Error("Type should be 'string'")
				}
				if m["description"] != "A test string" {
					t.Error("Description not preserved")
				}
			},
		},
		{
			name: "Object with properties",
			input: &parser.Schema{
				Type: "object",
				Properties: map[string]*parser.Schema{
					"name": {Type: "string"},
					"age":  {Type: "integer"},
				},
				Required: []string{"name"},
			},
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("Result should be a map")
				}
				props, ok := m["properties"].(map[string]interface{})
				if !ok {
					t.Fatal("Should have properties")
				}
				if len(props) != 2 {
					t.Error("Should have 2 properties")
				}
				required, ok := m["required"].([]string)
				if !ok || len(required) != 1 || required[0] != "name" {
					t.Error("Required field not correct")
				}
			},
		},
		{
			name: "Array schema",
			input: &parser.Schema{
				Type: "array",
				Items: &parser.Schema{
					Type: "string",
				},
			},
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("Result should be a map")
				}
				if m["type"] != "array" {
					t.Error("Type should be 'array'")
				}
				items, ok := m["items"].(map[string]interface{})
				if !ok {
					t.Fatal("Should have items")
				}
				if items["type"] != "string" {
					t.Error("Items type should be 'string'")
				}
			},
		},
		{
			name: "Enum schema",
			input: &parser.Schema{
				Type: "string",
				Enum: []interface{}{"red", "green", "blue"},
			},
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("Result should be a map")
				}
				enum, ok := m["enum"].([]interface{})
				if !ok || len(enum) != 3 {
					t.Error("Enum not preserved correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.convertSchema(tt.input)
			tt.validate(t, result)
		})
	}
}

func TestGenerateDescription(t *testing.T) {
	gen := NewMCPGenerator(&parser.OpenAPISpec{})

	tests := []struct {
		name     string
		op       *parser.Operation
		method   string
		path     string
		expected string
	}{
		{
			name: "With summary only",
			op: &parser.Operation{
				Summary: "Get all pets",
			},
			method:   "get",
			path:     "/pets",
			expected: "Get all pets",
		},
		{
			name: "With summary and description",
			op: &parser.Operation{
				Summary:     "Get all pets",
				Description: "Returns a list of all pets in the system",
			},
			method:   "get",
			path:     "/pets",
			expected: "Get all pets\nReturns a list of all pets in the system",
		},
		{
			name:     "No summary or description",
			op:       &parser.Operation{},
			method:   "get",
			path:     "/pets",
			expected: "GET /pets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.generateDescription(tt.op, tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("Expected '%s', got: '%s'", tt.expected, result)
			}
		})
	}
}
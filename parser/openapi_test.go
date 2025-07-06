package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOpenAPISpec(t *testing.T) {
	// Test with Petstore spec
	spec, err := ParseOpenAPISpec("../examples/petstore.yaml")
	if err != nil {
		t.Fatalf("Failed to parse Petstore spec: %v", err)
	}

	// Validate basic structure
	if spec.OpenAPI == "" {
		t.Error("OpenAPI version should not be empty")
	}

	if spec.Info.Title == "" {
		t.Error("Info.Title should not be empty")
	}

	if len(spec.Paths) == 0 {
		t.Error("Paths should not be empty")
	}

	// Check specific paths exist
	expectedPaths := []string{"/pet", "/pet/{petId}", "/store/order", "/user"}
	for _, path := range expectedPaths {
		if _, ok := spec.Paths[path]; !ok {
			t.Errorf("Expected path %s not found", path)
		}
	}

	// Check operations
	petPath := spec.Paths["/pet"]
	if petPath.Post == nil {
		t.Error("Expected POST /pet operation")
	}

	if petPath.Post.OperationID != "addPet" {
		t.Errorf("Expected operationId 'addPet', got: %s", petPath.Post.OperationID)
	}

	// Check components/schemas
	if spec.Components == nil || spec.Components.Schemas == nil {
		t.Error("Expected components with schemas")
	}

	if _, ok := spec.Components.Schemas["Pet"]; !ok {
		t.Error("Expected 'Pet' schema in components")
	}
}

func TestGetOperations(t *testing.T) {
	pathItem := &PathItem{
		Get: &Operation{
			OperationID: "getTest",
		},
		Post: &Operation{
			OperationID: "postTest",
		},
		Put: &Operation{
			OperationID: "putTest",
		},
	}

	ops := pathItem.GetOperations()

	if len(ops) != 3 {
		t.Errorf("Expected 3 operations, got: %d", len(ops))
	}

	if ops["get"].OperationID != "getTest" {
		t.Error("GET operation not correctly mapped")
	}

	if ops["post"].OperationID != "postTest" {
		t.Error("POST operation not correctly mapped")
	}

	if ops["put"].OperationID != "putTest" {
		t.Error("PUT operation not correctly mapped")
	}
}

func TestParseInvalidFile(t *testing.T) {
	_, err := ParseOpenAPISpec("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestParseInvalidYAML(t *testing.T) {
	// Create a temporary invalid YAML file
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.yaml")
	
	err := os.WriteFile(invalidFile, []byte("invalid: yaml: content:"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = ParseOpenAPISpec(invalidFile)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}
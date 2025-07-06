package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type OpenAPISpec struct {
	OpenAPI string                     `yaml:"openapi"`
	Info    Info                       `yaml:"info"`
	Servers []Server                   `yaml:"servers"`
	Paths   map[string]PathItem        `yaml:"paths"`
	Components *Components             `yaml:"components,omitempty"`
}

type Info struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

type Server struct {
	URL         string `yaml:"url"`
	Description string `yaml:"description"`
}

type PathItem struct {
	Get     *Operation `yaml:"get,omitempty"`
	Post    *Operation `yaml:"post,omitempty"`
	Put     *Operation `yaml:"put,omitempty"`
	Delete  *Operation `yaml:"delete,omitempty"`
	Patch   *Operation `yaml:"patch,omitempty"`
	Options *Operation `yaml:"options,omitempty"`
	Head    *Operation `yaml:"head,omitempty"`
}

type Operation struct {
	OperationID string                 `yaml:"operationId"`
	Summary     string                 `yaml:"summary"`
	Description string                 `yaml:"description"`
	Parameters  []Parameter            `yaml:"parameters,omitempty"`
	RequestBody *RequestBody           `yaml:"requestBody,omitempty"`
	Responses   map[string]Response    `yaml:"responses"`
	Tags        []string               `yaml:"tags,omitempty"`
}

type Parameter struct {
	Name        string      `yaml:"name"`
	In          string      `yaml:"in"`
	Description string      `yaml:"description"`
	Required    bool        `yaml:"required"`
	Schema      *Schema     `yaml:"schema"`
}

type RequestBody struct {
	Description string               `yaml:"description"`
	Required    bool                 `yaml:"required"`
	Content     map[string]MediaType `yaml:"content"`
}

type MediaType struct {
	Schema *Schema `yaml:"schema"`
}

type Response struct {
	Description string               `yaml:"description"`
	Content     map[string]MediaType `yaml:"content,omitempty"`
}

type Schema struct {
	Type        string              `yaml:"type"`
	Format      string              `yaml:"format,omitempty"`
	Properties  map[string]*Schema  `yaml:"properties,omitempty"`
	Items       *Schema             `yaml:"items,omitempty"`
	Required    []string            `yaml:"required,omitempty"`
	Ref         string              `yaml:"$ref,omitempty"`
	Enum        []interface{}       `yaml:"enum,omitempty"`
	Description string              `yaml:"description,omitempty"`
}

type Components struct {
	Schemas map[string]*Schema `yaml:"schemas,omitempty"`
}

func ParseOpenAPISpec(filePath string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var spec OpenAPISpec
	err = yaml.Unmarshal(data, &spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &spec, nil
}

func (p *PathItem) GetOperations() map[string]*Operation {
	ops := make(map[string]*Operation)
	if p.Get != nil {
		ops["get"] = p.Get
	}
	if p.Post != nil {
		ops["post"] = p.Post
	}
	if p.Put != nil {
		ops["put"] = p.Put
	}
	if p.Delete != nil {
		ops["delete"] = p.Delete
	}
	if p.Patch != nil {
		ops["patch"] = p.Patch
	}
	if p.Options != nil {
		ops["options"] = p.Options
	}
	if p.Head != nil {
		ops["head"] = p.Head
	}
	return ops
}
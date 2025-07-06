package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"specmill/server"
)

func main() {
	var specPath string
	flag.StringVar(&specPath, "spec", "", "Path to OpenAPI spec file (YAML)")
	flag.Parse()

	if specPath == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -spec <openapi-spec.yaml>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	srv, err := server.NewMCPServer(specPath)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
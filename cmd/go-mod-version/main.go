// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/mod/modfile"
)

// Map tool names to Go import paths (same mapping as get-tool-version.sh)
var toolPaths = map[string]string{
	"go-ctrf-json-reporter": "github.com/ctrf-io/go-ctrf-json-reporter",
	"go-junit-report":       "github.com/jstemmer/go-junit-report/v2",
	"gotestsum":             "gotest.tools/gotestsum",
	"svu":                   "github.com/caarlos0/svu",
	"go-mod-version":        "github.com/go-openapi/gh-actions/cmd/go-mod-version",
	"go-monorepo-detector":  "github.com/go-openapi/gh-actions/cmd/go-monorepo-detector",
	"gh-workflow-waiter":    "github.com/go-openapi/gh-actions/cmd/gh-workflow-waiter",
}

func main() {
	tool := flag.String("tool", "", "Tool name to look up")
	gomodPath := flag.String("gomod", "go.mod", "Path to go.mod file")
	flag.Parse()

	if *tool == "" {
		fmt.Fprintf(os.Stderr, "Error: -tool flag required\n")
		fmt.Fprintf(os.Stderr, "Usage: go-mod-version -tool <toolname> [-gomod <path>]\n")
		os.Exit(1)
	}

	version, err := getVersion(*gomodPath, *tool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(version)
}

func getVersion(gomodPath, toolName string) (string, error) {
	importPath, ok := toolPaths[toolName]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	data, err := os.ReadFile(gomodPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", gomodPath, err)
	}

	mod, err := modfile.Parse(gomodPath, data, nil)
	if err != nil {
		return "", fmt.Errorf("failed to parse go.mod: %w", err)
	}

	// Look for exact match or module with version suffix (e.g., /v2, /v3)
	for _, req := range mod.Require {
		if req.Mod.Path == importPath ||
			strings.HasPrefix(req.Mod.Path, importPath+"/") {
			return req.Mod.Version, nil
		}
	}

	return "", fmt.Errorf("version not found for %s (%s) in go.mod",
		toolName, importPath)
}

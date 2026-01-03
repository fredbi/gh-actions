// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name        string
		gomodData   string
		toolName    string
		wantVersion string
		wantErr     bool
	}{
		{
			name: "valid gotestsum lookup",
			gomodData: `module github.com/go-openapi/gh-actions

go 1.23

require (
	gotest.tools/gotestsum v1.13.0
)
`,
			toolName:    "gotestsum",
			wantVersion: "v1.13.0",
			wantErr:     false,
		},
		{
			name: "valid svu lookup with version suffix",
			gomodData: `module github.com/go-openapi/gh-actions

go 1.23

require (
	github.com/caarlos0/svu/v3 v3.3.0
)
`,
			toolName:    "svu",
			wantVersion: "v3.3.0",
			wantErr:     false,
		},
		{
			name: "valid go-junit-report with v2 suffix",
			gomodData: `module github.com/go-openapi/gh-actions

go 1.23

require (
	github.com/jstemmer/go-junit-report/v2 v2.1.0
)
`,
			toolName:    "go-junit-report",
			wantVersion: "v2.1.0",
			wantErr:     false,
		},
		{
			name: "valid go-ctrf-json-reporter",
			gomodData: `module github.com/go-openapi/gh-actions

go 1.23

require (
	github.com/ctrf-io/go-ctrf-json-reporter v0.0.15
)
`,
			toolName:    "go-ctrf-json-reporter",
			wantVersion: "v0.0.15",
			wantErr:     false,
		},
		{
			name: "unknown tool",
			gomodData: `module github.com/go-openapi/gh-actions

go 1.23

require (
	gotest.tools/gotestsum v1.13.0
)
`,
			toolName: "nonexistent-tool",
			wantErr:  true,
		},
		{
			name: "tool not in go.mod",
			gomodData: `module github.com/go-openapi/gh-actions

go 1.23

require (
	github.com/some/other v1.0.0
)
`,
			toolName: "gotestsum",
			wantErr:  true,
		},
		{
			name: "multiple tools in go.mod",
			gomodData: `module github.com/go-openapi/gh-actions

go 1.23

require (
	github.com/caarlos0/svu/v3 v3.3.0
	github.com/ctrf-io/go-ctrf-json-reporter v0.0.15
	github.com/jstemmer/go-junit-report/v2 v2.1.0
	gotest.tools/gotestsum v1.13.0
)
`,
			toolName:    "svu",
			wantVersion: "v3.3.0",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary go.mod file
			tmpDir := t.TempDir()
			gomodPath := filepath.Join(tmpDir, "go.mod")

			if err := os.WriteFile(gomodPath, []byte(tt.gomodData), 0600); err != nil {
				t.Fatalf("Failed to create test go.mod: %v", err)
			}

			// Test getVersion function
			gotVersion, err := getVersion(gomodPath, tt.toolName)

			if (err != nil) != tt.wantErr {
				t.Errorf("getVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && gotVersion != tt.wantVersion {
				t.Errorf("getVersion() = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}

func TestGetVersion_MissingFile(t *testing.T) {
	_, err := getVersion("/nonexistent/go.mod", "gotestsum")
	if err == nil {
		t.Error("Expected error for missing go.mod, got nil")
	}
}

func TestGetVersion_MalformedGoMod(t *testing.T) {
	tmpDir := t.TempDir()
	gomodPath := filepath.Join(tmpDir, "go.mod")

	malformedData := `this is not valid go.mod syntax{{{`
	if err := os.WriteFile(gomodPath, []byte(malformedData), 0600); err != nil {
		t.Fatalf("Failed to create test go.mod: %v", err)
	}

	_, err := getVersion(gomodPath, "gotestsum")
	if err == nil {
		t.Error("Expected error for malformed go.mod, got nil")
	}
}

func TestToolPaths(t *testing.T) {
	// Verify all expected tools are mapped
	expectedTools := []string{
		"go-ctrf-json-reporter",
		"go-junit-report",
		"gotestsum",
		"svu",
		"go-mod-version",
		"go-monorepo-detector",
		"gh-workflow-waiter",
	}

	for _, tool := range expectedTools {
		if _, ok := toolPaths[tool]; !ok {
			t.Errorf("Tool %s not found in toolPaths map", tool)
		}
	}
}

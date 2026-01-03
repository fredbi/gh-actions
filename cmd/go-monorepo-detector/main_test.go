// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectModules_SingleModule(t *testing.T) {
	// Create a temporary directory with a single module
	tmpDir := t.TempDir()

	// Create go.mod
	gomod := `module github.com/example/single

go 1.23
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(gomod), 0600); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	output, err := detectModules()
	if err != nil {
		t.Fatalf("detectModules() failed: %v", err)
	}

	// Verify single module results
	if output.ModulesCount != 1 {
		t.Errorf("ModulesCount = %d, want 1", output.ModulesCount)
	}

	if output.IsMonorepo {
		t.Error("IsMonorepo = true, want false for single module")
	}

	if output.RootModule != "github.com/example/single" {
		t.Errorf("RootModule = %s, want github.com/example/single", output.RootModule)
	}

	if len(output.Modules) != 1 {
		t.Errorf("len(Modules) = %d, want 1", len(output.Modules))
	}

	if len(output.RelativeNames) != 1 || output.RelativeNames[0] != "" {
		t.Errorf("RelativeNames[0] = %q, want empty string for root", output.RelativeNames[0])
	}

	if output.BashRelNames != "{root}" {
		t.Errorf("BashRelNames = %q, want {root}", output.BashRelNames)
	}
}

func TestPrintOutput_JSON(t *testing.T) {
	output := &Output{
		Modules: []Module{
			{Name: "github.com/example/test", Path: "/tmp/test"},
		},
		RootModule:      "github.com/example/test",
		Paths:           []string{"/tmp/test"},
		Names:           []string{"github.com/example/test"},
		RelativeNames:   []string{""},
		BashPaths:       "/tmp/test",
		BashSubpaths:    "./...",
		BashRelNames:    "{root}",
		ModulesCount:    1,
		IsMonorepo:      false,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printOutput(output, "json")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("printOutput() failed: %v", err)
	}

	// Read captured output
	var result Output
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON output: %v", err)
	}

	if result.RootModule != output.RootModule {
		t.Errorf("JSON RootModule = %s, want %s", result.RootModule, output.RootModule)
	}

	if result.ModulesCount != output.ModulesCount {
		t.Errorf("JSON ModulesCount = %d, want %d", result.ModulesCount, output.ModulesCount)
	}
}

func TestMustJSON(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "string slice",
			input: []string{"a", "b", "c"},
			want:  `["a","b","c"]`,
		},
		{
			name:  "empty slice",
			input: []string{},
			want:  `[]`,
		},
		{
			name: "module slice",
			input: []Module{
				{Name: "mod1", Path: "/path1"},
			},
			want: `[{"name":"mod1","path":"/path1"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustJSON(tt.input)
			if got != tt.want {
				t.Errorf("mustJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

// Note: Testing monorepo scenarios requires go.work files which are more complex to set up
// These would be covered by integration tests in CI

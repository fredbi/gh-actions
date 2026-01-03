// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Module represents a Go module with its name and path
type Module struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Output contains all the module information in various formats
type Output struct {
	Modules         []Module `json:"modules"`
	RootModule      string   `json:"root_module"`
	Paths           []string `json:"paths"`
	Names           []string `json:"names"`
	RelativeNames   []string `json:"relative_names"`
	BashPaths       string   `json:"bash_paths"`
	BashSubpaths    string   `json:"bash_subpaths"`
	BashRelNames    string   `json:"bash_relative_names"`
	ModulesCount    int      `json:"modules_count"`
	IsMonorepo      bool     `json:"is_monorepo"`
}

func main() {
	format := flag.String("format", "json", "Output format: json, github-actions, or individual field names")
	flag.Parse()

	output, err := detectModules()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := printOutput(output, *format); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func detectModules() (*Output, error) {
	// Get root module
	rootCmd := exec.Command("go", "list")
	rootOutput, err := rootCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get root module: %w", err)
	}
	root := strings.TrimSpace(string(rootOutput))

	// Get all modules with their paths
	listCmd := exec.Command("go", "list", "-m", "-f", `{"name":{{ printf "%q" .Path }},"path":{{ printf "%q" .Dir }}}`)
	listOutput, err := listCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}

	// Parse module information
	lines := strings.Split(strings.TrimSpace(string(listOutput)), "\n")
	modules := make([]Module, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		var mod Module
		if err := json.Unmarshal([]byte(line), &mod); err != nil {
			return nil, fmt.Errorf("failed to parse module info: %w", err)
		}
		modules = append(modules, mod)
	}

	// Build output structure
	output := &Output{
		Modules:      modules,
		RootModule:   root,
		ModulesCount: len(modules),
		IsMonorepo:   len(modules) > 1,
	}

	// Extract paths and names
	output.Paths = make([]string, len(modules))
	output.Names = make([]string, len(modules))
	output.RelativeNames = make([]string, len(modules))

	bashPaths := make([]string, 0, len(modules))
	bashSubpaths := make([]string, 0, len(modules))
	bashRelNames := make([]string, 0, len(modules))

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	for i, mod := range modules {
		output.Paths[i] = mod.Path
		output.Names[i] = mod.Name

		// Calculate relative name
		relName := strings.TrimPrefix(mod.Name, root)
		relName = strings.TrimPrefix(relName, "/")
		output.RelativeNames[i] = relName

		// Bash outputs
		bashPaths = append(bashPaths, mod.Path)

		// Subpaths for go test (relative path with /...)
		relPath := strings.TrimPrefix(mod.Path, cwd)
		if relPath == mod.Path {
			// Not under cwd, use absolute
			relPath = mod.Path
		} else if !strings.HasPrefix(relPath, ".") {
			relPath = "." + relPath
		}
		bashSubpaths = append(bashSubpaths, relPath+"/...")

		// Bash relative names (root module shown as {root})
		if relName == "" {
			bashRelNames = append(bashRelNames, "{root}")
		} else {
			bashRelNames = append(bashRelNames, relName)
		}
	}

	output.BashPaths = strings.Join(bashPaths, " ")
	output.BashSubpaths = strings.Join(bashSubpaths, " ")
	output.BashRelNames = strings.Join(bashRelNames, " ")

	return output, nil
}

func printOutput(output *Output, format string) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)

	case "github-actions":
		// Output in GitHub Actions format (GITHUB_OUTPUT)
		fmt.Printf("modules=%s\n", mustJSON(output.Modules))
		fmt.Printf("root-module=%s\n", output.RootModule)
		fmt.Printf("paths=%s\n", mustJSON(output.Paths))
		fmt.Printf("names=%s\n", mustJSON(output.Names))
		fmt.Printf("relative-names=%s\n", mustJSON(output.RelativeNames))
		fmt.Printf("bash-paths=%s\n", output.BashPaths)
		fmt.Printf("bash-subpaths=%s\n", output.BashSubpaths)
		fmt.Printf("bash-relative-names=%s\n", output.BashRelNames)
		fmt.Printf("modules-count=%d\n", output.ModulesCount)
		fmt.Printf("is-monorepo=%t\n", output.IsMonorepo)
		return nil

	case "modules":
		return json.NewEncoder(os.Stdout).Encode(output.Modules)
	case "root-module":
		fmt.Println(output.RootModule)
		return nil
	case "paths":
		return json.NewEncoder(os.Stdout).Encode(output.Paths)
	case "names":
		return json.NewEncoder(os.Stdout).Encode(output.Names)
	case "relative-names":
		return json.NewEncoder(os.Stdout).Encode(output.RelativeNames)
	case "bash-paths":
		fmt.Println(output.BashPaths)
		return nil
	case "bash-subpaths":
		fmt.Println(output.BashSubpaths)
		return nil
	case "bash-relative-names":
		fmt.Println(output.BashRelNames)
		return nil
	case "modules-count":
		fmt.Println(output.ModulesCount)
		return nil
	case "is-monorepo":
		fmt.Println(output.IsMonorepo)
		return nil

	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func mustJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}

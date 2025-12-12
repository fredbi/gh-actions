# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This repository contains reusable GitHub Actions for go-openapi workflows. The actions install vetted versions of Go testing and release tools from binary releases rather than using `go install`, allowing for pinned CI dependencies and vulnerability scanning.

## Key Architecture

### Tool Installation Strategy

Tools are installed via downloaded binary releases, not `go install`. The version tracking mechanism works as follows:

1. **Version Resolution**: `get-tool-version.sh` extracts versions from `go.mod` by mapping tool names to Go import paths
2. **Dependency Tracking**: `release_tracker.go` is a dummy Go package that imports all tracked tools, allowing Dependabot to monitor and propose updates
3. **Vulnerability Scanning**: Updates require passing vulnerability scans on the source repositories before approval

### Directory Structure

- `install/*/action.yml`: Individual tool installer actions (gotestsum, go-junit-report, go-ctrf-json-reporter, svu)
- `ci-jobs/wait-pending-jobs/action.yml`: Reusable action to wait for all workflow runs on a PR's head SHA to complete
- `action.yml`: Composite action that installs all tools at once
- `get-tool-version.sh`: Bash script to resolve tool versions from go.mod without requiring Go installed

### wait-pending-jobs Action

Solves a timing issue where auto-merge can trigger before non-required jobs (like coverage uploads) complete, causing those jobs to fail when the branch gets deleted. Uses `gh` CLI to poll workflow runs via GitHub API until all jobs reach a terminal state.

## Development Commands

### Testing the Actions

Run the test workflow locally or verify via GitHub Actions:
```bash
# The test.yml workflow installs all tools and verifies they work
# across ubuntu-latest, macos-latest, and windows-latest
```

### Linting

The repository uses actionlint for workflow validation:
```bash
# Lint is done in CI via raven-actions/actionlint
# Only lints .github/**/*.yml files (not action.yml in subdirectories)
```

### Version Updates

When adding or updating a tool:
1. Update `go.mod` to include the new version
2. Add import to `release_tracker.go` if it's a new tool
3. Add case to `get-tool-version.sh` with the correct Go import path
4. Create/update the corresponding `install/*/action.yml`

Tool import paths in `get-tool-version.sh`:
- `go-ctrf-json-reporter` → `github.com/ctrf-io/go-ctrf-json-reporter`
- `go-junit-report` → `github.com/jstemmer/go-junit-report/v2`
- `gotestsum` → `gotest.tools/gotestsum`
- `svu` → `github.com/caarlos0/svu`

## Release Process

Maintainers cut releases by:
1. Running the bump-release.yml workflow, OR
2. Pushing a semver tag (signed tags preferred; tag message prepends release notes)

## Contributing

- Branch naming: `fix/XXX-something` for bugs, `feat/XXX-something` for features
- Commit messages: Follow conventional commit style (e.g., "fix:", "feat:", "ci:")
- Sign commits with: `Signed-off-by: Your Name <email@example.com>`
- Squash commits into logical units before merge
- Go version required: old stable (latest minor - 1)

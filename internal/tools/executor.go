// Package tools provides code execution capabilities for Notion tools.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Executor executes code from Notion code blocks.
type Executor struct {
	timeout   time.Duration
	languages map[string]bool
}

// NewExecutor creates a new code executor.
func NewExecutor(timeout time.Duration, languages string) *Executor {
	langMap := make(map[string]bool)
	for _, lang := range strings.Split(languages, ",") {
		lang = strings.TrimSpace(lang)
		if lang != "" {
			langMap[lang] = true
		}
	}
	return &Executor{
		timeout:   timeout,
		languages: langMap,
	}
}

// ExecutionResult represents the result of code execution.
type ExecutionResult struct {
	Output   string
	Error    string
	ExitCode int
}

// Execute executes code in the specified language.
func (e *Executor) Execute(ctx context.Context, language, code string, input any) (*ExecutionResult, error) {
	// Check if language is allowed
	if !e.isLanguageAllowed(language) {
		return nil, fmt.Errorf("language %q is not allowed", language)
	}

	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	var output string
	var exitCode int
	var err error

	switch language {
	case "bash", "sh":
		output, exitCode, err = e.executeBash(ctx, code, input)
	case "python", "py":
		output, exitCode, err = e.executePython(ctx, code, input)
	case "js", "javascript":
		output, exitCode, err = e.executeNode(ctx, code, input)
	case "ts", "typescript":
		output, exitCode, err = e.executeTsNode(ctx, code, input)
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	result := &ExecutionResult{
		Output:   output,
		ExitCode: exitCode,
	}
	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}

// isLanguageAllowed checks if a language is in the allowed list.
func (e *Executor) isLanguageAllowed(language string) bool {
	if len(e.languages) == 0 {
		// If no languages specified, allow all
		return true
	}
	return e.languages[language]
}

// executeBash executes bash code.
func (e *Executor) executeBash(ctx context.Context, code string, input any) (string, int, error) {
	cmd := exec.CommandContext(ctx, "bash", "-c", code)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(output), exitErr.ExitCode(), nil
		}
		return string(output), -1, err
	}
	return string(output), 0, nil
}

// executePython executes python code.
func (e *Executor) executePython(ctx context.Context, code string, input any) (string, int, error) {
	cmd := exec.CommandContext(ctx, "python3", "-c", code)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(output), exitErr.ExitCode(), nil
		}
		return string(output), -1, err
	}
	return string(output), 0, nil
}

// executeNode executes JavaScript code.
func (e *Executor) executeNode(ctx context.Context, code string, input any) (string, int, error) {
	cmd := exec.CommandContext(ctx, "node", "-e", code)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(output), exitErr.ExitCode(), nil
		}
		return string(output), -1, err
	}
	return string(output), 0, nil
}

func (e *Executor) executeTsNode(ctx context.Context, code string, input any) (string, int, error) {
	jsonInput, err := json.Marshal(input)
	if err != nil {
		return "", -1, fmt.Errorf("failed to marshal input: %w", err)
	}
	// Escape the JSON string for safe embedding in JavaScript string literal
	// Escape backslashes first, then single quotes
	jsonStr := strings.ReplaceAll(string(jsonInput), `\`, `\\`)
	jsonStr = strings.ReplaceAll(jsonStr, `'`, `\'`)
	// Use JSON.parse to safely parse the JSON string, and console.log to output the result
	codeRun := fmt.Sprintf("%s\n console.log(JSON.stringify(handle(JSON.parse('%s'))));", code, jsonStr)
	cmd := exec.CommandContext(ctx, "npx", "ts-node", "--compiler-options",
		`{"module":"commonjs","moduleResolution":"node"}`, "-e", codeRun)
	cmd.Env = append(cmd.Env, "NODE_TLS_REJECT_UNAUTHORIZED=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(output), exitErr.ExitCode(), nil
		}
		return string(output), -1, err
	}
	return string(output), 0, nil
}

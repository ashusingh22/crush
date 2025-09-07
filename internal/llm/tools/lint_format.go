package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/crush/internal/language"
	"github.com/charmbracelet/crush/internal/permission"
)

type LintFormatParams struct {
	Action string   `json:"action"` // "lint", "format", "both"
	Files  []string `json:"files,omitempty"`
	Language string `json:"language,omitempty"` // Optional override
}

type LintFormatResult struct {
	Action   string                   `json:"action"`
	Success  bool                     `json:"success"`
	Results  map[string]interface{}   `json:"results"`
	Errors   []string                 `json:"errors,omitempty"`
	Language string                   `json:"language"`
}

type lintFormatTool struct {
	permissions permission.Service
	workingDir  string
}

const LintFormatToolName = "lint_format"

func NewLintFormatTool(permissions permission.Service, workingDir string) BaseTool {
	return &lintFormatTool{
		permissions: permissions,
		workingDir:  workingDir,
	}
}

func (t *lintFormatTool) Info() ToolInfo {
	return ToolInfo{
		Name:        LintFormatToolName,
		Description: "Lint and format code files using language-specific tools. Supports Go, Python, JavaScript/TypeScript, PHP, Rust, and Java.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"enum":        []string{"lint", "format", "both"},
					"description": "Action to perform: lint code, format code, or both",
				},
				"files": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Specific files to process (optional - if not provided, will process entire project)",
				},
				"language": map[string]any{
					"type":        "string",
					"description": "Override language detection (optional)",
				},
			},
			"required": []string{"action"},
		},
	}
}

func (t *lintFormatTool) Name() string {
	return LintFormatToolName
}

func (t *lintFormatTool) Run(ctx context.Context, params ToolCall) (ToolResponse, error) {
	var lintParams LintFormatParams
	if err := json.Unmarshal([]byte(params.Input), &lintParams); err != nil {
		return NewTextErrorResponse("Invalid parameters"), nil
	}

	// Detect language if not provided
	languageName := lintParams.Language
	var langConfig *language.SupportedLanguage
	
	if languageName == "" {
		detectedLang, detectedConfig, err := language.DetectLanguage(t.workingDir)
		if err != nil {
			return NewTextErrorResponse(fmt.Sprintf("Failed to detect language: %v", err)), nil
		}
		languageName = detectedLang
		langConfig = detectedConfig
	} else {
		config := language.DefaultLanguageConfig()
		if lang, exists := config.Languages[languageName]; exists {
			langConfig = &lang
		} else {
			return NewTextErrorResponse(fmt.Sprintf("Unsupported language: %s", languageName)), nil
		}
	}

	result := &LintFormatResult{
		Action:   lintParams.Action,
		Language: languageName,
		Results:  make(map[string]interface{}),
		Success:  true,
	}

	// Request permission for potentially modifying operations
	if lintParams.Action == "format" || lintParams.Action == "both" {
		sessionID, _ := GetContextValues(ctx)
		if sessionID != "" && params.ID != "" {
			granted := t.permissions.Request(permission.CreatePermissionRequest{
				SessionID:   sessionID,
				ToolCallID:  params.ID,
				ToolName:    LintFormatToolName,
				Action:      "format",
				Path:        t.workingDir,
				Description: fmt.Sprintf("Format %s code files", languageName),
			})
			if !granted {
				return NewTextErrorResponse("Permission denied to format files"), nil
			}
		}
	}

	// Perform linting if requested
	if lintParams.Action == "lint" || lintParams.Action == "both" {
		if langConfig.LintCommand != "" {
			lintResult, err := t.runLinter(langConfig.LintCommand, lintParams.Files)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Lint error: %v", err))
				result.Success = false
			}
			result.Results["lint"] = lintResult
		} else {
			result.Results["lint"] = "No linter configured for " + languageName
		}
	}

	// Perform formatting if requested
	if lintParams.Action == "format" || lintParams.Action == "both" {
		if langConfig.FormatCommand != "" {
			formatResult, err := t.runFormatter(langConfig.FormatCommand, lintParams.Files)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Format error: %v", err))
				result.Success = false
			}
			result.Results["format"] = formatResult
		} else {
			result.Results["format"] = "No formatter configured for " + languageName
		}
	}

	output, _ := json.Marshal(result)
	return NewTextResponse(string(output)), nil
}

func (t *lintFormatTool) runLinter(command string, files []string) (map[string]interface{}, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty lint command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	if len(files) > 0 {
		cmd.Args = append(cmd.Args, files...)
	}
	cmd.Dir = t.workingDir

	output, err := cmd.CombinedOutput()
	
	result := map[string]interface{}{
		"command": command,
		"output":  string(output),
		"success": err == nil,
	}

	if err != nil {
		result["error"] = err.Error()
	}

	return result, nil
}

func (t *lintFormatTool) runFormatter(command string, files []string) (map[string]interface{}, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty format command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	if len(files) > 0 {
		cmd.Args = append(cmd.Args, files...)
	}
	cmd.Dir = t.workingDir

	output, err := cmd.CombinedOutput()
	
	result := map[string]interface{}{
		"command": command,
		"output":  string(output),
		"success": err == nil,
	}

	if err != nil {
		result["error"] = err.Error()
	}

	return result, nil
}

// Enhanced language detection for specific files
func (t *lintFormatTool) detectLanguageForFiles(files []string) (string, *language.SupportedLanguage, error) {
	if len(files) == 0 {
		return language.DetectLanguage(t.workingDir)
	}

	// Use the first file's extension to determine language
	ext := filepath.Ext(files[0])
	langName, langConfig := language.GetLanguageByExtension(ext)
	if langName == "" {
		return "", nil, fmt.Errorf("could not detect language for file extension: %s", ext)
	}

	return langName, langConfig, nil
}
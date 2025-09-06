package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/permission"
)

type AnalyzeParams struct {
	Path string `json:"path"`
	Type string `json:"type"` // "structure", "complexity", "dependencies", "patterns"
}

type AnalysisResult struct {
	Type        string                 `json:"type"`
	Summary     string                 `json:"summary"`
	Details     map[string]interface{} `json:"details"`
	Suggestions []string              `json:"suggestions"`
	Timestamp   time.Time             `json:"timestamp"`
}

type analyzeTool struct {
	permissions permission.Service
	workingDir  string
}

const AnalyzeToolName = "analyze"

func NewAnalyzeTool(permissions permission.Service, workingDir string) BaseTool {
	return &analyzeTool{
		permissions: permissions,
		workingDir:  workingDir,
	}
}

func (t *analyzeTool) Info() ToolInfo {
	return ToolInfo{
		Name:        AnalyzeToolName,
		Description: "Analyze code structure, complexity, dependencies, and patterns without LLM calls. Supports Go, JavaScript, Python, and general file analysis.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to file or directory to analyze",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Type of analysis: structure, complexity, dependencies, patterns",
					"enum":        []string{"structure", "complexity", "dependencies", "patterns"},
				},
			},
			"required": []string{"path", "type"},
		},
		Required: []string{"path", "type"},
	}
}

func (t *analyzeTool) Name() string {
	return AnalyzeToolName
}

func (t *analyzeTool) Run(ctx context.Context, params ToolCall) (ToolResponse, error) {
	var analyzeParams AnalyzeParams
	if err := json.Unmarshal([]byte(params.Input), &analyzeParams); err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Failed to parse parameters: %v", err)), nil
	}

	if analyzeParams.Path == "" {
		return NewTextErrorResponse("Path parameter is required"), nil
	}

	path := analyzeParams.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(t.workingDir, path)
	}

	// Check permissions
	sessionID, _ := GetContextValues(ctx)
	if !t.permissions.Request(permission.CreatePermissionRequest{
		SessionID:   sessionID,
		ToolCallID:  params.ID,
		ToolName:    AnalyzeToolName,
		Description: fmt.Sprintf("Analyze %s (%s)", path, analyzeParams.Type),
		Action:      "analyze:" + analyzeParams.Type,
		Path:        path,
		Params:      analyzeParams,
	}) {
		return NewTextErrorResponse("Permission denied"), nil
	}

	// Perform analysis based on type
	result, err := t.performAnalysis(path, analyzeParams.Type)
	if err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Analysis failed: %v", err)), nil
	}

	// Format result
	output := t.formatAnalysisResult(result)
	return NewTextResponse(output), nil
}

func (t *analyzeTool) performAnalysis(path, analysisType string) (*AnalysisResult, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path: %w", err)
	}

	result := &AnalysisResult{
		Type:      analysisType,
		Details:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	if stat.IsDir() {
		return t.analyzeDirectory(path, analysisType, result)
	}
	return t.analyzeFile(path, analysisType, result)
}

func (t *analyzeTool) analyzeDirectory(dirPath, analysisType string, result *AnalysisResult) (*AnalysisResult, error) {
	switch analysisType {
	case "structure":
		return t.analyzeDirectoryStructure(dirPath, result)
	case "complexity":
		return t.analyzeDirectoryComplexity(dirPath, result)
	case "dependencies":
		return t.analyzeDirectoryDependencies(dirPath, result)
	case "patterns":
		return t.analyzeDirectoryPatterns(dirPath, result)
	default:
		return nil, fmt.Errorf("unsupported analysis type: %s", analysisType)
	}
}

func (t *analyzeTool) analyzeFile(filePath, analysisType string, result *AnalysisResult) (*AnalysisResult, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	switch analysisType {
	case "structure":
		return t.analyzeFileStructure(filePath, ext, result)
	case "complexity":
		return t.analyzeFileComplexity(filePath, ext, result)
	case "dependencies":
		return t.analyzeFileDependencies(filePath, ext, result)
	case "patterns":
		return t.analyzeFilePatterns(filePath, ext, result)
	default:
		return nil, fmt.Errorf("unsupported analysis type: %s", analysisType)
	}
}

func (t *analyzeTool) analyzeDirectoryStructure(dirPath string, result *AnalysisResult) (*AnalysisResult, error) {
	structure := make(map[string]interface{})
	fileCount := 0
	dirCount := 0
	languages := make(map[string]int)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			dirCount++
		} else {
			fileCount++
			ext := strings.ToLower(filepath.Ext(path))
			if ext != "" {
				languages[ext]++
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	structure["total_files"] = fileCount
	structure["total_directories"] = dirCount
	structure["languages"] = languages

	result.Summary = fmt.Sprintf("Directory contains %d files and %d directories", fileCount, dirCount)
	result.Details = structure

	// Add suggestions
	if fileCount > 1000 {
		result.Suggestions = append(result.Suggestions, "Large project - consider using focused analysis on specific subdirectories")
	}
	if len(languages) > 5 {
		result.Suggestions = append(result.Suggestions, "Multi-language project - consider language-specific analysis")
	}

	return result, nil
}

func (t *analyzeTool) analyzeFileStructure(filePath, ext string, result *AnalysisResult) (*AnalysisResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	structure := make(map[string]interface{})
	
	lines := strings.Split(string(content), "\n")
	structure["line_count"] = len(lines)
	structure["character_count"] = len(content)
	structure["file_size_bytes"] = len(content)

	// Language-specific analysis
	switch ext {
	case ".go":
		return t.analyzeGoFileStructure(filePath, content, result)
	case ".js", ".ts":
		return t.analyzeJSFileStructure(content, result)
	case ".py":
		return t.analyzePythonFileStructure(content, result)
	default:
		result.Summary = fmt.Sprintf("File has %d lines and %d characters", len(lines), len(content))
		result.Details = structure
	}

	return result, nil
}

func (t *analyzeTool) analyzeGoFileStructure(filePath string, content []byte, result *AnalysisResult) (*AnalysisResult, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	structure := make(map[string]interface{})
	
	// Count different elements
	var functions, types, vars, consts int
	var imports []string

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			functions++
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				types++
			case token.VAR:
				vars++
			case token.CONST:
				consts++
			case token.IMPORT:
				for _, spec := range d.Specs {
					if importSpec, ok := spec.(*ast.ImportSpec); ok {
						importPath := strings.Trim(importSpec.Path.Value, `"`)
						imports = append(imports, importPath)
					}
				}
			}
		}
	}

	structure["package"] = node.Name.Name
	structure["functions"] = functions
	structure["types"] = types
	structure["variables"] = vars
	structure["constants"] = consts
	structure["imports"] = imports
	structure["import_count"] = len(imports)

	result.Summary = fmt.Sprintf("Go file with %d functions, %d types, %d imports", functions, types, len(imports))
	result.Details = structure

	// Add suggestions
	if functions > 20 {
		result.Suggestions = append(result.Suggestions, "Consider splitting large file into multiple files")
	}
	if len(imports) > 15 {
		result.Suggestions = append(result.Suggestions, "High number of imports - consider dependency analysis")
	}

	return result, nil
}

func (t *analyzeTool) analyzeJSFileStructure(content []byte, result *AnalysisResult) (*AnalysisResult, error) {
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	
	structure := make(map[string]interface{})
	
	// Count functions, classes, imports
	functionCount := strings.Count(contentStr, "function ") + strings.Count(contentStr, "=> ")
	classCount := strings.Count(contentStr, "class ")
	importCount := strings.Count(contentStr, "import ") + strings.Count(contentStr, "require(")

	structure["functions"] = functionCount
	structure["classes"] = classCount
	structure["imports"] = importCount
	structure["lines"] = len(lines)

	result.Summary = fmt.Sprintf("JavaScript file with %d functions, %d classes, %d imports", functionCount, classCount, importCount)
	result.Details = structure

	return result, nil
}

func (t *analyzeTool) analyzePythonFileStructure(content []byte, result *AnalysisResult) (*AnalysisResult, error) {
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	
	structure := make(map[string]interface{})
	
	// Count functions, classes, imports
	functionCount := strings.Count(contentStr, "def ")
	classCount := strings.Count(contentStr, "class ")
	importCount := strings.Count(contentStr, "import ") + strings.Count(contentStr, "from ")

	structure["functions"] = functionCount
	structure["classes"] = classCount
	structure["imports"] = importCount
	structure["lines"] = len(lines)

	result.Summary = fmt.Sprintf("Python file with %d functions, %d classes, %d imports", functionCount, classCount, importCount)
	result.Details = structure

	return result, nil
}

func (t *analyzeTool) analyzeFileComplexity(filePath, ext string, result *AnalysisResult) (*AnalysisResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	complexity := make(map[string]interface{})

	// Basic complexity metrics
	lines := strings.Split(contentStr, "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	// Count control flow statements
	ifCount := strings.Count(contentStr, "if ") + strings.Count(contentStr, "if(")
	forCount := strings.Count(contentStr, "for ") + strings.Count(contentStr, "for(")
	whileCount := strings.Count(contentStr, "while ") + strings.Count(contentStr, "while(")
	switchCount := strings.Count(contentStr, "switch ") + strings.Count(contentStr, "switch(")

	complexity["lines_of_code"] = nonEmptyLines
	complexity["if_statements"] = ifCount
	complexity["loops"] = forCount + whileCount
	complexity["switch_statements"] = switchCount
	complexity["cyclomatic_complexity"] = ifCount + forCount + whileCount + switchCount + 1

	result.Summary = fmt.Sprintf("Cyclomatic complexity: %d", complexity["cyclomatic_complexity"])
	result.Details = complexity

	// Add suggestions based on complexity
	if cc := complexity["cyclomatic_complexity"].(int); cc > 10 {
		result.Suggestions = append(result.Suggestions, "High cyclomatic complexity - consider refactoring")
	}
	if nonEmptyLines > 300 {
		result.Suggestions = append(result.Suggestions, "Large file - consider splitting into smaller modules")
	}

	return result, nil
}

func (t *analyzeTool) analyzeDirectoryComplexity(dirPath string, result *AnalysisResult) (*AnalysisResult, error) {
	// Analyze complexity across all files in directory
	totalComplexity := 0
	fileCount := 0
	
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".go" || ext == ".js" || ext == ".py" || ext == ".ts" {
			fileResult, err := t.analyzeFileComplexity(path, ext, &AnalysisResult{Details: make(map[string]interface{})})
			if err == nil {
				if cc, ok := fileResult.Details["cyclomatic_complexity"].(int); ok {
					totalComplexity += cc
					fileCount++
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	avgComplexity := 0
	if fileCount > 0 {
		avgComplexity = totalComplexity / fileCount
	}

	result.Details["total_complexity"] = totalComplexity
	result.Details["average_complexity"] = avgComplexity
	result.Details["analyzed_files"] = fileCount
	result.Summary = fmt.Sprintf("Average complexity: %d across %d files", avgComplexity, fileCount)

	if avgComplexity > 15 {
		result.Suggestions = append(result.Suggestions, "High average complexity - consider code refactoring")
	}

	return result, nil
}

func (t *analyzeTool) analyzeFileDependencies(filePath, ext string, result *AnalysisResult) (*AnalysisResult, error) {
	// Implement dependency analysis for different file types
	result.Summary = "Dependency analysis not yet implemented for this file type"
	return result, nil
}

func (t *analyzeTool) analyzeDirectoryDependencies(dirPath string, result *AnalysisResult) (*AnalysisResult, error) {
	// Implement directory-wide dependency analysis
	result.Summary = "Directory dependency analysis not yet implemented"
	return result, nil
}

func (t *analyzeTool) analyzeFilePatterns(filePath, ext string, result *AnalysisResult) (*AnalysisResult, error) {
	// Implement pattern analysis (design patterns, anti-patterns, etc.)
	result.Summary = "Pattern analysis not yet implemented for this file type"
	return result, nil
}

func (t *analyzeTool) analyzeDirectoryPatterns(dirPath string, result *AnalysisResult) (*AnalysisResult, error) {
	// Implement directory-wide pattern analysis
	result.Summary = "Directory pattern analysis not yet implemented"
	return result, nil
}

func (t *analyzeTool) formatAnalysisResult(result *AnalysisResult) string {
	var output strings.Builder
	
	output.WriteString(fmt.Sprintf("# %s Analysis\n\n", strings.Title(result.Type)))
	output.WriteString(fmt.Sprintf("**Summary:** %s\n\n", result.Summary))
	
	if len(result.Details) > 0 {
		output.WriteString("## Details\n\n")
		for key, value := range result.Details {
			output.WriteString(fmt.Sprintf("- **%s:** %v\n", strings.Title(strings.ReplaceAll(key, "_", " ")), value))
		}
		output.WriteString("\n")
	}
	
	if len(result.Suggestions) > 0 {
		output.WriteString("## Suggestions\n\n")
		for _, suggestion := range result.Suggestions {
			output.WriteString(fmt.Sprintf("- %s\n", suggestion))
		}
		output.WriteString("\n")
	}
	
	output.WriteString(fmt.Sprintf("*Analysis completed at %s*", result.Timestamp.Format(time.RFC3339)))
	
	return output.String()
}
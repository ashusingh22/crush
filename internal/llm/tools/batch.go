package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/permission"
)

type BatchParams struct {
	Operations []BatchOperation `json:"operations"`
	Parallel   bool             `json:"parallel"`
}

type BatchOperation struct {
	Type   string                 `json:"type"`   // "file_search", "text_replace", "file_copy", "dir_analysis"
	Params map[string]interface{} `json:"params"`
}

type BatchResult struct {
	OperationIndex int         `json:"operation_index"`
	Type           string      `json:"type"`
	Success        bool        `json:"success"`
	Result         interface{} `json:"result"`
	Error          string      `json:"error,omitempty"`
	Duration       string      `json:"duration"`
}

type batchTool struct {
	permissions permission.Service
	workingDir  string
}

const BatchToolName = "batch"

func NewBatchTool(permissions permission.Service, workingDir string) BaseTool {
	return &batchTool{
		permissions: permissions,
		workingDir:  workingDir,
	}
}

func (t *batchTool) Info() ToolInfo {
	return ToolInfo{
		Name:        BatchToolName,
		Description: "Execute multiple operations in batch to reduce API calls. Supports file operations, searches, and analysis tasks.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"operations": map[string]any{
					"type": "array",
					"description": "Array of operations to execute",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"type": map[string]any{
								"type":        "string",
								"description": "Operation type: file_search, text_replace, file_copy, dir_analysis, pattern_find",
								"enum":        []string{"file_search", "text_replace", "file_copy", "dir_analysis", "pattern_find"},
							},
							"params": map[string]any{
								"type":        "object",
								"description": "Operation-specific parameters",
							},
						},
						"required": []string{"type", "params"},
					},
				},
				"parallel": map[string]any{
					"type":        "boolean",
					"description": "Whether to execute operations in parallel (default: false)",
					"default":     false,
				},
			},
			"required": []string{"operations"},
		},
		Required: []string{"operations"},
	}
}

func (t *batchTool) Name() string {
	return BatchToolName
}

func (t *batchTool) Run(ctx context.Context, params ToolCall) (ToolResponse, error) {
	var batchParams BatchParams
	if err := json.Unmarshal([]byte(params.Input), &batchParams); err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Failed to parse parameters: %v", err)), nil
	}

	if len(batchParams.Operations) == 0 {
		return NewTextErrorResponse("No operations specified"), nil
	}

	sessionID, _ := GetContextValues(ctx)

	// Check permissions for batch operations
	if !t.permissions.Request(permission.CreatePermissionRequest{
		SessionID:   sessionID,
		ToolCallID:  params.ID,
		ToolName:    BatchToolName,
		Description: fmt.Sprintf("Execute %d batch operations", len(batchParams.Operations)),
		Action:      "execute_batch",
		Path:        t.workingDir,
		Params:      batchParams,
	}) {
		return NewTextErrorResponse("Permission denied"), nil
	}

	var results []BatchResult

	if batchParams.Parallel {
		results = t.executeParallel(ctx, batchParams.Operations)
	} else {
		results = t.executeSequential(ctx, batchParams.Operations)
	}

	// Format results
	output := t.formatBatchResults(results)
	return NewTextResponse(output), nil
}

func (t *batchTool) executeSequential(ctx context.Context, operations []BatchOperation) []BatchResult {
	results := make([]BatchResult, len(operations))

	for i, op := range operations {
		start := time.Now()
		result, err := t.executeOperation(ctx, op)
		duration := time.Since(start)

		results[i] = BatchResult{
			OperationIndex: i,
			Type:           op.Type,
			Success:        err == nil,
			Result:         result,
			Duration:       duration.String(),
		}

		if err != nil {
			results[i].Error = err.Error()
		}
	}

	return results
}

func (t *batchTool) executeParallel(ctx context.Context, operations []BatchOperation) []BatchResult {
	results := make([]BatchResult, len(operations))
	resultChan := make(chan struct {
		index  int
		result BatchResult
	}, len(operations))

	// Start all operations
	for i, op := range operations {
		go func(index int, operation BatchOperation) {
			start := time.Now()
			result, err := t.executeOperation(ctx, operation)
			duration := time.Since(start)

			batchResult := BatchResult{
				OperationIndex: index,
				Type:           operation.Type,
				Success:        err == nil,
				Result:         result,
				Duration:       duration.String(),
			}

			if err != nil {
				batchResult.Error = err.Error()
			}

			resultChan <- struct {
				index  int
				result BatchResult
			}{index: index, result: batchResult}
		}(i, op)
	}

	// Collect results
	for i := 0; i < len(operations); i++ {
		res := <-resultChan
		results[res.index] = res.result
	}

	return results
}

func (t *batchTool) executeOperation(ctx context.Context, op BatchOperation) (interface{}, error) {
	switch op.Type {
	case "file_search":
		return t.executeFileSearch(op.Params)
	case "text_replace":
		return t.executeTextReplace(op.Params)
	case "file_copy":
		return t.executeFileCopy(op.Params)
	case "dir_analysis":
		return t.executeDirAnalysis(op.Params)
	case "pattern_find":
		return t.executePatternFind(op.Params)
	default:
		return nil, fmt.Errorf("unsupported operation type: %s", op.Type)
	}
}

func (t *batchTool) executeFileSearch(params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter required for file_search")
	}

	searchPath := t.workingDir
	if path, ok := params["path"].(string); ok {
		if !filepath.IsAbs(path) {
			searchPath = filepath.Join(t.workingDir, path)
		} else {
			searchPath = path
		}
	}

	var matches []string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		// Check if filename matches query
		if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(query)) {
			relPath, _ := filepath.Rel(t.workingDir, path)
			matches = append(matches, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"query":        query,
		"search_path":  searchPath,
		"matches":      matches,
		"match_count":  len(matches),
	}, nil
}

func (t *batchTool) executeTextReplace(params map[string]interface{}) (interface{}, error) {
	filePath, ok := params["file"].(string)
	if !ok {
		return nil, fmt.Errorf("file parameter required for text_replace")
	}

	oldText, ok := params["old_text"].(string)
	if !ok {
		return nil, fmt.Errorf("old_text parameter required for text_replace")
	}

	newText, ok := params["new_text"].(string)
	if !ok {
		return nil, fmt.Errorf("new_text parameter required for text_replace")
	}

	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(t.workingDir, filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)
	replacedContent := strings.ReplaceAll(originalContent, oldText, newText)
	replacementCount := strings.Count(originalContent, oldText)

	if replacementCount == 0 {
		return map[string]interface{}{
			"file":             filePath,
			"replacements":     0,
			"modified":         false,
		}, nil
	}

	err = os.WriteFile(filePath, []byte(replacedContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"file":             filePath,
		"old_text":         oldText,
		"new_text":         newText,
		"replacements":     replacementCount,
		"modified":         true,
	}, nil
}

func (t *batchTool) executeFileCopy(params map[string]interface{}) (interface{}, error) {
	source, ok := params["source"].(string)
	if !ok {
		return nil, fmt.Errorf("source parameter required for file_copy")
	}

	destination, ok := params["destination"].(string)
	if !ok {
		return nil, fmt.Errorf("destination parameter required for file_copy")
	}

	if !filepath.IsAbs(source) {
		source = filepath.Join(t.workingDir, source)
	}
	if !filepath.IsAbs(destination) {
		destination = filepath.Join(t.workingDir, destination)
	}

	sourceContent, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}

	// Create destination directory if needed
	destDir := filepath.Dir(destination)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	err = os.WriteFile(destination, sourceContent, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write destination file: %w", err)
	}

	sourceInfo, _ := os.Stat(source)
	return map[string]interface{}{
		"source":       source,
		"destination":  destination,
		"size_bytes":   len(sourceContent),
		"copied":       true,
		"source_info":  sourceInfo,
	}, nil
}

func (t *batchTool) executeDirAnalysis(params map[string]interface{}) (interface{}, error) {
	analysisPath := t.workingDir
	if path, ok := params["path"].(string); ok {
		if !filepath.IsAbs(path) {
			analysisPath = filepath.Join(t.workingDir, path)
		} else {
			analysisPath = path
		}
	}

	analysis := map[string]interface{}{
		"path":             analysisPath,
		"total_files":      0,
		"total_dirs":       0,
		"total_size":       int64(0),
		"file_types":       make(map[string]int),
		"largest_files":    []map[string]interface{}{},
	}

	var largestFiles []map[string]interface{}

	err := filepath.Walk(analysisPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			analysis["total_dirs"] = analysis["total_dirs"].(int) + 1
		} else {
			analysis["total_files"] = analysis["total_files"].(int) + 1
			analysis["total_size"] = analysis["total_size"].(int64) + info.Size()

			// Track file types
			ext := strings.ToLower(filepath.Ext(path))
			if ext == "" {
				ext = "[no extension]"
			}
			fileTypes := analysis["file_types"].(map[string]int)
			fileTypes[ext]++

			// Track largest files
			relPath, _ := filepath.Rel(analysisPath, path)
			fileInfo := map[string]interface{}{
				"path":      relPath,
				"size":      info.Size(),
				"modified":  info.ModTime(),
			}

			largestFiles = append(largestFiles, fileInfo)
			
			// Keep only top 10 largest files
			if len(largestFiles) > 10 {
				// Simple bubble sort to keep largest
				for i := 0; i < len(largestFiles)-1; i++ {
					for j := 0; j < len(largestFiles)-i-1; j++ {
						if largestFiles[j]["size"].(int64) < largestFiles[j+1]["size"].(int64) {
							largestFiles[j], largestFiles[j+1] = largestFiles[j+1], largestFiles[j]
						}
					}
				}
				largestFiles = largestFiles[:10]
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	analysis["largest_files"] = largestFiles
	return analysis, nil
}

func (t *batchTool) executePatternFind(params map[string]interface{}) (interface{}, error) {
	pattern, ok := params["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("pattern parameter required for pattern_find")
	}

	searchPath := t.workingDir
	if path, ok := params["path"].(string); ok {
		if !filepath.IsAbs(path) {
			searchPath = filepath.Join(t.workingDir, path)
		} else {
			searchPath = path
		}
	}

	fileExtensions := []string{".go", ".js", ".ts", ".py", ".java", ".cpp", ".c", ".h"}
	if exts, ok := params["extensions"].([]interface{}); ok {
		fileExtensions = nil
		for _, ext := range exts {
			if extStr, ok := ext.(string); ok {
				fileExtensions = append(fileExtensions, extStr)
			}
		}
	}

	var matches []map[string]interface{}

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Check if file extension is in our list
		ext := strings.ToLower(filepath.Ext(path))
		extMatch := false
		for _, allowedExt := range fileExtensions {
			if ext == allowedExt {
				extMatch = true
				break
			}
		}

		if !extMatch {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			if strings.Contains(strings.ToLower(line), strings.ToLower(pattern)) {
				relPath, _ := filepath.Rel(searchPath, path)
				matches = append(matches, map[string]interface{}{
					"file":       relPath,
					"line":       lineNum + 1,
					"content":    strings.TrimSpace(line),
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"pattern":     pattern,
		"search_path": searchPath,
		"matches":     matches,
		"match_count": len(matches),
	}, nil
}

func (t *batchTool) formatBatchResults(results []BatchResult) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("# Batch Operation Results\n\n"))
	output.WriteString(fmt.Sprintf("Executed %d operations\n\n", len(results)))

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	output.WriteString(fmt.Sprintf("**Success Rate:** %d/%d (%.1f%%)\n\n", 
		successCount, len(results), float64(successCount)/float64(len(results))*100))

	for _, result := range results {
		output.WriteString(fmt.Sprintf("## Operation %d: %s\n", result.OperationIndex+1, result.Type))
		output.WriteString(fmt.Sprintf("**Status:** %s | **Duration:** %s\n\n", 
			map[bool]string{true: "✅ Success", false: "❌ Failed"}[result.Success], result.Duration))

		if !result.Success {
			output.WriteString(fmt.Sprintf("**Error:** %s\n\n", result.Error))
		} else {
			// Format result based on operation type
			switch result.Type {
			case "file_search":
				if resultMap, ok := result.Result.(map[string]interface{}); ok {
					output.WriteString(fmt.Sprintf("Found %v matches for query '%v'\n\n", 
						resultMap["match_count"], resultMap["query"]))
				}
			case "text_replace":
				if resultMap, ok := result.Result.(map[string]interface{}); ok {
					output.WriteString(fmt.Sprintf("Made %v replacements in %v\n\n", 
						resultMap["replacements"], filepath.Base(resultMap["file"].(string))))
				}
			case "file_copy":
				if resultMap, ok := result.Result.(map[string]interface{}); ok {
					output.WriteString(fmt.Sprintf("Copied %v bytes to %v\n\n", 
						resultMap["size_bytes"], filepath.Base(resultMap["destination"].(string))))
				}
			case "dir_analysis":
				if resultMap, ok := result.Result.(map[string]interface{}); ok {
					output.WriteString(fmt.Sprintf("Analyzed: %v files, %v directories, %v bytes total\n\n", 
						resultMap["total_files"], resultMap["total_dirs"], resultMap["total_size"]))
				}
			case "pattern_find":
				if resultMap, ok := result.Result.(map[string]interface{}); ok {
					output.WriteString(fmt.Sprintf("Found %v matches for pattern '%v'\n\n", 
						resultMap["match_count"], resultMap["pattern"]))
				}
			}
		}
	}

	return output.String()
}
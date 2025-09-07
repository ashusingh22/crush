package language

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SupportedLanguage represents a programming language with its configuration
type SupportedLanguage struct {
	Name        string   `json:"name"`
	Extensions  []string `json:"extensions"`
	LSPCommand  string   `json:"lsp_command,omitempty"`
	LintCommand string   `json:"lint_command,omitempty"`
	FormatCommand string `json:"format_command,omitempty"`
	BuildCommand string  `json:"build_command,omitempty"`
	TestCommand string   `json:"test_command,omitempty"`
	ProjectFiles []string `json:"project_files,omitempty"`
}

// LanguageConfig holds the configuration for all supported languages
type LanguageConfig struct {
	Languages map[string]SupportedLanguage `json:"languages"`
}

// DefaultLanguageConfig returns the default configuration for supported languages
func DefaultLanguageConfig() *LanguageConfig {
	return &LanguageConfig{
		Languages: map[string]SupportedLanguage{
			"go": {
				Name:         "Go",
				Extensions:   []string{".go"},
				LSPCommand:   "gopls",
				LintCommand:  "golangci-lint run",
				FormatCommand: "gofmt -w",
				BuildCommand: "go build",
				TestCommand:  "go test",
				ProjectFiles: []string{"go.mod", "go.sum"},
			},
			"python": {
				Name:         "Python",
				Extensions:   []string{".py", ".pyx"},
				LSPCommand:   "pylsp",
				LintCommand:  "pylint",
				FormatCommand: "black",
				BuildCommand: "python -m py_compile",
				TestCommand:  "python -m pytest",
				ProjectFiles: []string{"setup.py", "pyproject.toml", "requirements.txt", "Pipfile", "poetry.lock"},
			},
			"javascript": {
				Name:         "JavaScript",
				Extensions:   []string{".js", ".jsx", ".mjs"},
				LSPCommand:   "typescript-language-server --stdio",
				LintCommand:  "eslint",
				FormatCommand: "prettier --write",
				BuildCommand: "npm run build",
				TestCommand:  "npm test",
				ProjectFiles: []string{"package.json", "package-lock.json", "yarn.lock", "pnpm-lock.yaml"},
			},
			"typescript": {
				Name:         "TypeScript",
				Extensions:   []string{".ts", ".tsx"},
				LSPCommand:   "typescript-language-server --stdio",
				LintCommand:  "eslint",
				FormatCommand: "prettier --write",
				BuildCommand: "tsc",
				TestCommand:  "npm test",
				ProjectFiles: []string{"tsconfig.json", "package.json", "package-lock.json"},
			},
			"php": {
				Name:         "PHP",
				Extensions:   []string{".php", ".phtml"},
				LSPCommand:   "intelephense --stdio",
				LintCommand:  "phpcs",
				FormatCommand: "phpcbf",
				BuildCommand: "php -l",
				TestCommand:  "phpunit",
				ProjectFiles: []string{"composer.json", "composer.lock", "phpunit.xml"},
			},
			"rust": {
				Name:         "Rust",
				Extensions:   []string{".rs"},
				LSPCommand:   "rust-analyzer",
				LintCommand:  "cargo clippy",
				FormatCommand: "cargo fmt",
				BuildCommand: "cargo build",
				TestCommand:  "cargo test",
				ProjectFiles: []string{"Cargo.toml", "Cargo.lock"},
			},
			"java": {
				Name:         "Java",
				Extensions:   []string{".java"},
				LSPCommand:   "jdtls",
				LintCommand:  "checkstyle",
				FormatCommand: "google-java-format",
				BuildCommand: "javac",
				TestCommand:  "mvn test",
				ProjectFiles: []string{"pom.xml", "build.gradle", "build.gradle.kts"},
			},
		},
	}
}

// DetectLanguage detects the primary language of a project based on files in the directory
func DetectLanguage(projectPath string) (string, *SupportedLanguage, error) {
	config := DefaultLanguageConfig()
	
	// Check for project-specific files first
	for langName, lang := range config.Languages {
		for _, projectFile := range lang.ProjectFiles {
			if _, err := os.Stat(filepath.Join(projectPath, projectFile)); err == nil {
				return langName, &lang, nil
			}
		}
	}
	
	// Count files by extension
	extensionCounts := make(map[string]int)
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}
		if info.IsDir() {
			// Skip common directories
			dirName := info.Name()
			if strings.HasPrefix(dirName, ".") || 
			   dirName == "node_modules" ||
			   dirName == "vendor" ||
			   dirName == "target" ||
			   dirName == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}
		
		ext := strings.ToLower(filepath.Ext(path))
		if ext != "" {
			extensionCounts[ext]++
		}
		return nil
	})
	
	if err != nil {
		return "", nil, fmt.Errorf("failed to walk directory: %w", err)
	}
	
	// Find the language with the most files
	var bestLang string
	var bestCount int
	var bestConfig *SupportedLanguage
	
	for langName, lang := range config.Languages {
		count := 0
		for _, ext := range lang.Extensions {
			count += extensionCounts[ext]
		}
		if count > bestCount {
			bestCount = count
			bestLang = langName
			langCopy := lang
			bestConfig = &langCopy
		}
	}
	
	if bestLang == "" {
		return "", nil, fmt.Errorf("could not detect language for project")
	}
	
	return bestLang, bestConfig, nil
}

// GetLanguageByExtension returns the language configuration for a given file extension
func GetLanguageByExtension(ext string) (string, *SupportedLanguage) {
	config := DefaultLanguageConfig()
	ext = strings.ToLower(ext)
	
	for langName, lang := range config.Languages {
		for _, langExt := range lang.Extensions {
			if langExt == ext {
				langCopy := lang
				return langName, &langCopy
			}
		}
	}
	
	return "", nil
}

// SaveLanguageConfig saves the language configuration to a file
func (lc *LanguageConfig) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(lc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	return os.WriteFile(filename, data, 0644)
}

// LoadLanguageConfig loads the language configuration from a file
func LoadLanguageConfig(filename string) (*LanguageConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config LanguageConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &config, nil
}
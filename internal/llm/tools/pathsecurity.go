package tools

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
)

// ValidatePathSecurity validates and sanitizes file paths to prevent directory traversal attacks
func ValidatePathSecurity(requestedPath, workingDir string) (string, error) {
	// Sanitize the path
	sanitizedPath := filepath.Clean(requestedPath)
	
	// Check for obvious path traversal attempts
	if strings.Contains(sanitizedPath, "..") {
		slog.Warn("🚨 SECURITY: Path traversal attempt blocked",
			"requested_path", requestedPath,
			"sanitized_path", sanitizedPath,
		)
		return "", fmt.Errorf("path traversal not allowed: %s", requestedPath)
	}

	// Get absolute working directory
	workingDirAbs, err := filepath.Abs(workingDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve working directory: %w", err)
	}

	var finalPath string
	if filepath.IsAbs(sanitizedPath) {
		// For absolute paths, ensure they are within the working directory or a safe location
		finalPath = sanitizedPath
	} else {
		// For relative paths, join with working directory
		finalPath = filepath.Join(workingDirAbs, sanitizedPath)
	}

	// Get absolute final path
	finalPathAbs, err := filepath.Abs(finalPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve final path: %w", err)
	}

	// Ensure the final path is within the working directory
	rel, err := filepath.Rel(workingDirAbs, finalPathAbs)
	if err != nil {
		return "", fmt.Errorf("failed to compute relative path: %w", err)
	}

	// Check if the path escapes the working directory
	if strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, "/") {
		slog.Warn("🚨 SECURITY: Path outside working directory blocked",
			"requested_path", requestedPath,
			"working_dir", workingDirAbs,
			"resolved_path", finalPathAbs,
			"relative_path", rel,
		)
		return "", fmt.Errorf("path resolves outside working directory: %s", requestedPath)
	}

	slog.Debug("Path validation successful",
		"requested_path", requestedPath,
		"final_path", finalPathAbs,
		"relative_path", rel,
	)

	return finalPathAbs, nil
}

// ValidatePathSecurityRelative is like ValidatePathSecurity but returns a path relative to workingDir
func ValidatePathSecurityRelative(requestedPath, workingDir string) (string, error) {
	finalPath, err := ValidatePathSecurity(requestedPath, workingDir)
	if err != nil {
		return "", err
	}

	workingDirAbs, err := filepath.Abs(workingDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve working directory: %w", err)
	}

	rel, err := filepath.Rel(workingDirAbs, finalPath)
	if err != nil {
		return "", fmt.Errorf("failed to compute relative path: %w", err)
	}

	return rel, nil
}
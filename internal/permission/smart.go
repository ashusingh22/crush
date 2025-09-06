package permission

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SmartPermissionPattern represents a learned permission pattern
type SmartPermissionPattern struct {
	ToolName      string    `json:"tool_name"`
	Action        string    `json:"action"`
	PathPattern   string    `json:"path_pattern"`
	ApprovalCount int       `json:"approval_count"`
	DenialCount   int       `json:"denial_count"`
	LastUsed      time.Time `json:"last_used"`
	Confidence    float64   `json:"confidence"`
	AutoApprove   bool      `json:"auto_approve"`
}

// SmartPermissionService extends the basic permission service with learning capabilities
type SmartPermissionService struct {
	Service
	patterns            map[string]*SmartPermissionPattern
	patternsMu          sync.RWMutex
	learningFile        string
	enabled             bool
	confidenceThreshold float64
}

// NewSmartPermissionService creates an enhanced permission service with learning
func NewSmartPermissionService(baseService Service, workingDir string, enabled bool) *SmartPermissionService {
	sps := &SmartPermissionService{
		Service:             baseService,
		patterns:            make(map[string]*SmartPermissionPattern),
		learningFile:        filepath.Join(workingDir, ".crush", "permission_patterns.json"),
		enabled:             enabled,
		confidenceThreshold: 0.8, // Auto-approve when confidence >= 80%
	}

	if enabled {
		sps.loadPatterns()
	}

	return sps
}

// Request overrides the base Request method to add smart learning
func (s *SmartPermissionService) Request(opts CreatePermissionRequest) bool {
	if !s.enabled {
		return s.Service.Request(opts)
	}

	// Check if we have a learned pattern for this request
	if s.shouldAutoApprove(opts) {
		slog.Debug("Auto-approving based on learned pattern",
			"tool", opts.ToolName,
			"action", opts.Action,
			"path", opts.Path,
		)
		return true
	}

	// Fall back to regular permission check
	approved := s.Service.Request(opts)

	// Learn from the user's decision
	s.learnFromDecision(opts, approved)

	return approved
}

// shouldAutoApprove checks if the request matches a high-confidence pattern
func (s *SmartPermissionService) shouldAutoApprove(opts CreatePermissionRequest) bool {
	s.patternsMu.RLock()
	defer s.patternsMu.RUnlock()

	key := s.getPatternKey(opts.ToolName, opts.Action, opts.Path)
	pattern, exists := s.patterns[key]

	if !exists {
		return false
	}

	// Update last used time
	pattern.LastUsed = time.Now()

	// Check if pattern is confident enough and set to auto-approve
	return pattern.AutoApprove && pattern.Confidence >= s.confidenceThreshold
}

// learnFromDecision records the user's decision to improve future predictions
func (s *SmartPermissionService) learnFromDecision(opts CreatePermissionRequest, approved bool) {
	s.patternsMu.Lock()
	defer s.patternsMu.Unlock()

	key := s.getPatternKey(opts.ToolName, opts.Action, opts.Path)
	pattern, exists := s.patterns[key]

	if !exists {
		pattern = &SmartPermissionPattern{
			ToolName:    opts.ToolName,
			Action:      opts.Action,
			PathPattern: s.generalizePattern(opts.Path),
		}
		s.patterns[key] = pattern
	}

	// Update counts
	if approved {
		pattern.ApprovalCount++
	} else {
		pattern.DenialCount++
	}

	pattern.LastUsed = time.Now()

	// Calculate confidence and auto-approval eligibility
	s.updatePatternConfidence(pattern)

	slog.Debug("Learned from permission decision",
		"tool", opts.ToolName,
		"action", opts.Action,
		"approved", approved,
		"confidence", pattern.Confidence,
		"auto_approve", pattern.AutoApprove,
	)

	// Save patterns asynchronously
	go s.savePatterns()
}

// updatePatternConfidence calculates confidence and auto-approval eligibility
func (s *SmartPermissionService) updatePatternConfidence(pattern *SmartPermissionPattern) {
	total := pattern.ApprovalCount + pattern.DenialCount
	if total == 0 {
		pattern.Confidence = 0
		pattern.AutoApprove = false
		return
	}

	// Base confidence is approval rate
	approvalRate := float64(pattern.ApprovalCount) / float64(total)

	// Adjust confidence based on sample size (more samples = higher confidence)
	sampleSizeBonus := 1.0
	if total >= 5 {
		sampleSizeBonus = 1.1
	}
	if total >= 10 {
		sampleSizeBonus = 1.2
	}

	// Time decay for old patterns
	daysSinceLastUse := time.Since(pattern.LastUsed).Hours() / 24
	timeDecay := 1.0
	if daysSinceLastUse > 30 {
		timeDecay = 0.9 // Reduce confidence for old patterns
	}
	if daysSinceLastUse > 90 {
		timeDecay = 0.7
	}

	pattern.Confidence = approvalRate * sampleSizeBonus * timeDecay

	// Enable auto-approval for consistently approved actions
	pattern.AutoApprove = pattern.Confidence >= s.confidenceThreshold &&
		pattern.ApprovalCount >= 3 &&
		pattern.DenialCount == 0
}

// getPatternKey creates a unique key for permission patterns
func (s *SmartPermissionService) getPatternKey(toolName, action, path string) string {
	generalizedPath := s.generalizePattern(path)
	return fmt.Sprintf("%s:%s:%s", toolName, action, generalizedPath)
}

// generalizePattern creates a generalized pattern from a specific path
func (s *SmartPermissionService) generalizePattern(path string) string {
	// Convert absolute paths to relative patterns
	if filepath.IsAbs(path) {
		if rel, err := filepath.Rel(filepath.Dir(s.learningFile), path); err == nil {
			path = rel
		}
	}

	// Generalize common patterns
	patterns := []struct {
		pattern     string
		replacement string
	}{
		{`\d+`, "*"},          // Replace numbers with wildcards
		{`[a-f0-9]{8,}`, "*"}, // Replace long hex strings (IDs, hashes)
		{`\.tmp\w*`, ".tmp*"}, // Generalize temp files
		{`_\d+\.`, "_*."},     // Replace versioned files
	}

	generalized := path
	for _, p := range patterns {
		// Simple string replacement for basic patterns
		if strings.Contains(generalized, p.pattern) {
			generalized = strings.ReplaceAll(generalized, p.pattern, p.replacement)
		}
	}

	return generalized
}

// loadPatterns loads learned patterns from disk
func (s *SmartPermissionService) loadPatterns() {
	data, err := os.ReadFile(s.learningFile)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("Failed to load permission patterns", "error", err)
		}
		return
	}

	var patterns map[string]*SmartPermissionPattern
	if err := json.Unmarshal(data, &patterns); err != nil {
		slog.Warn("Failed to parse permission patterns", "error", err)
		return
	}

	s.patternsMu.Lock()
	s.patterns = patterns
	s.patternsMu.Unlock()

	slog.Debug("Loaded permission patterns", "count", len(patterns))
}

// savePatterns saves learned patterns to disk
func (s *SmartPermissionService) savePatterns() {
	s.patternsMu.RLock()
	patterns := make(map[string]*SmartPermissionPattern)
	for k, v := range s.patterns {
		patterns[k] = v
	}
	s.patternsMu.RUnlock()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.learningFile), 0755); err != nil {
		slog.Warn("Failed to create patterns directory", "error", err)
		return
	}

	data, err := json.MarshalIndent(patterns, "", "  ")
	if err != nil {
		slog.Warn("Failed to marshal permission patterns", "error", err)
		return
	}

	if err := os.WriteFile(s.learningFile, data, 0644); err != nil {
		slog.Warn("Failed to save permission patterns", "error", err)
		return
	}

	slog.Debug("Saved permission patterns", "count", len(patterns))
}

// GetLearningStats returns statistics about learned patterns
func (s *SmartPermissionService) GetLearningStats() map[string]interface{} {
	s.patternsMu.RLock()
	defer s.patternsMu.RUnlock()

	stats := map[string]interface{}{
		"enabled":                  s.enabled,
		"total_patterns":           len(s.patterns),
		"auto_approve_patterns":    0,
		"high_confidence_patterns": 0,
	}

	for _, pattern := range s.patterns {
		if pattern.AutoApprove {
			stats["auto_approve_patterns"] = stats["auto_approve_patterns"].(int) + 1
		}
		if pattern.Confidence >= s.confidenceThreshold {
			stats["high_confidence_patterns"] = stats["high_confidence_patterns"].(int) + 1
		}
	}

	return stats
}

// ClearLearning removes all learned patterns
func (s *SmartPermissionService) ClearLearning() error {
	s.patternsMu.Lock()
	s.patterns = make(map[string]*SmartPermissionPattern)
	s.patternsMu.Unlock()

	if err := os.Remove(s.learningFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	slog.Info("Cleared all learned permission patterns")
	return nil
}

// IsSafeOperation determines if an operation is generally safe to auto-approve
func (s *SmartPermissionService) IsSafeOperation(toolName, action string) bool {
	safeOperations := map[string][]string{
		"view":        {"read"},
		"ls":          {"list"},
		"grep":        {"search"},
		"glob":        {"search"},
		"analyze":     {"analyze:structure", "analyze:complexity", "analyze:dependencies", "analyze:patterns"},
		"batch":       {"execute_batch"}, // Batch operations with individual permission checks
		"diagnostics": {"get"},
	}

	if allowedActions, exists := safeOperations[toolName]; exists {
		for _, allowedAction := range allowedActions {
			if action == allowedAction {
				return true
			}
		}
	}

	return false
}

// SuggestAutoApproval suggests tools/actions that could be auto-approved based on patterns
func (s *SmartPermissionService) SuggestAutoApproval() []string {
	s.patternsMu.RLock()
	defer s.patternsMu.RUnlock()

	var suggestions []string

	for _, pattern := range s.patterns {
		if pattern.Confidence >= 0.9 && pattern.ApprovalCount >= 5 && pattern.DenialCount == 0 {
			suggestion := fmt.Sprintf("%s:%s (confidence: %.2f, used %d times)",
				pattern.ToolName, pattern.Action, pattern.Confidence, pattern.ApprovalCount)
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions
}

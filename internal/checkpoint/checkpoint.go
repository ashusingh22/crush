package checkpoint

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/permission"
)

// CheckpointService provides Git-based checkpoint functionality
type CheckpointService struct {
	workingDir  string
	permissions permission.Service
}

// Checkpoint represents a saved checkpoint
type Checkpoint struct {
	ID          string    `json:"id"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Hash        string    `json:"hash"`
	Branch      string    `json:"branch"`
	Files       []string  `json:"files"`
	IsStashed   bool      `json:"is_stashed"`
}

// CheckpointList holds multiple checkpoints
type CheckpointList struct {
	Checkpoints []Checkpoint `json:"checkpoints"`
}

// NewCheckpointService creates a new checkpoint service
func NewCheckpointService(workingDir string, permissions permission.Service) *CheckpointService {
	return &CheckpointService{
		workingDir:  workingDir,
		permissions: permissions,
	}
}

// CreateCheckpoint creates a new checkpoint by committing current changes
func (cs *CheckpointService) CreateCheckpoint(ctx context.Context, message string) (*Checkpoint, error) {
	// Check if we're in a git repository
	if !cs.isGitRepo() {
		return nil, fmt.Errorf("not in a git repository")
	}

	// Get current branch
	branch, err := cs.getCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check for uncommitted changes
	hasChanges, err := cs.hasUncommittedChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to check for changes: %w", err)
	}

	var checkpoint *Checkpoint

	if hasChanges {
		// Create checkpoint by stashing changes with a message
		stashMessage := fmt.Sprintf("crush-checkpoint: %s", message)
		if err := cs.runGitCommand("stash", "push", "-m", stashMessage, "--include-untracked"); err != nil {
			return nil, fmt.Errorf("failed to create stash: %w", err)
		}

		// Get the stash hash
		stashHash, err := cs.getLatestStashHash()
		if err != nil {
			return nil, fmt.Errorf("failed to get stash hash: %w", err)
		}

		checkpoint = &Checkpoint{
			ID:        fmt.Sprintf("stash-%d", time.Now().Unix()),
			Message:   message,
			Timestamp: time.Now(),
			Hash:      stashHash,
			Branch:    branch,
			IsStashed: true,
		}

		slog.Info("Created checkpoint via stash", "message", message, "hash", stashHash)
	} else {
		// No changes to checkpoint
		return nil, fmt.Errorf("no uncommitted changes to checkpoint")
	}

	return checkpoint, nil
}

// ListCheckpoints lists all available checkpoints (stashes and recent commits)
func (cs *CheckpointService) ListCheckpoints(ctx context.Context) (*CheckpointList, error) {
	if !cs.isGitRepo() {
		return nil, fmt.Errorf("not in a git repository")
	}

	var checkpoints []Checkpoint

	// Get stashes
	stashes, err := cs.getStashes()
	if err != nil {
		slog.Warn("Failed to get stashes", "error", err)
	} else {
		checkpoints = append(checkpoints, stashes...)
	}

	// Get recent commits (last 10)
	commits, err := cs.getRecentCommits(10)
	if err != nil {
		slog.Warn("Failed to get recent commits", "error", err)
	} else {
		checkpoints = append(checkpoints, commits...)
	}

	return &CheckpointList{Checkpoints: checkpoints}, nil
}

// RestoreCheckpoint restores a checkpoint by applying a stash or resetting to a commit
func (cs *CheckpointService) RestoreCheckpoint(ctx context.Context, checkpointID string) error {
	if !cs.isGitRepo() {
		return fmt.Errorf("not in a git repository")
	}

	// Request permission for potentially destructive operation
	sessionID, messageID := getContextValues(ctx)
	if sessionID != "" && messageID != "" {
		granted := cs.permissions.Request(permission.CreatePermissionRequest{
			SessionID:   sessionID,
			ToolCallID:  messageID,
			ToolName:    "checkpoint_restore",
			Action:      "restore",
			Path:        cs.workingDir,
			Description: fmt.Sprintf("Restore checkpoint %s (this will overwrite current changes)", checkpointID),
		})
		if !granted {
			return fmt.Errorf("permission denied to restore checkpoint")
		}
	}

	if strings.HasPrefix(checkpointID, "stash-") {
		// Restore from stash
		return cs.restoreFromStash(checkpointID)
	} else {
		// Restore from commit
		return cs.restoreFromCommit(checkpointID)
	}
}

// DeleteCheckpoint deletes a checkpoint (drops a stash)
func (cs *CheckpointService) DeleteCheckpoint(ctx context.Context, checkpointID string) error {
	if !cs.isGitRepo() {
		return fmt.Errorf("not in a git repository")
	}

	if strings.HasPrefix(checkpointID, "stash-") {
		// Find and drop the stash
		stashes, err := cs.getStashes()
		if err != nil {
			return fmt.Errorf("failed to get stashes: %w", err)
		}

		for i, stash := range stashes {
			if stash.ID == checkpointID {
				if err := cs.runGitCommand("stash", "drop", fmt.Sprintf("stash@{%d}", i)); err != nil {
					return fmt.Errorf("failed to drop stash: %w", err)
				}
				slog.Info("Deleted checkpoint", "id", checkpointID)
				return nil
			}
		}
		return fmt.Errorf("checkpoint not found: %s", checkpointID)
	}

	return fmt.Errorf("cannot delete commit checkpoints")
}

// isGitRepo checks if the current directory is a git repository
func (cs *CheckpointService) isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = cs.workingDir
	return cmd.Run() == nil
}

// getCurrentBranch gets the current git branch
func (cs *CheckpointService) getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = cs.workingDir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// hasUncommittedChanges checks if there are uncommitted changes
func (cs *CheckpointService) hasUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = cs.workingDir
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// runGitCommand runs a git command in the working directory
func (cs *CheckpointService) runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = cs.workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// getLatestStashHash gets the hash of the latest stash
func (cs *CheckpointService) getLatestStashHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "stash@{0}")
	cmd.Dir = cs.workingDir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getStashes gets all stashes as checkpoints
func (cs *CheckpointService) getStashes() ([]Checkpoint, error) {
	cmd := exec.Command("git", "stash", "list", "--format=%H|%gD|%gs|%at")
	cmd.Dir = cs.workingDir
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var checkpoints []Checkpoint
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for i, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}

		hash := parts[0]
		message := strings.TrimPrefix(parts[2], "On ")
		
		// Extract crush checkpoint message
		if strings.Contains(message, "crush-checkpoint:") {
			message = strings.TrimPrefix(message, "WIP on ")
			if idx := strings.Index(message, ": crush-checkpoint:"); idx != -1 {
				message = strings.TrimSpace(message[idx+len(": crush-checkpoint:"):])
			}
		}

		timestamp := time.Unix(parseUnixTimestamp(parts[3]), 0)

		checkpoints = append(checkpoints, Checkpoint{
			ID:        fmt.Sprintf("stash-%d", i),
			Message:   message,
			Timestamp: timestamp,
			Hash:      hash,
			IsStashed: true,
		})
	}

	return checkpoints, nil
}

// getRecentCommits gets recent commits as checkpoints
func (cs *CheckpointService) getRecentCommits(limit int) ([]Checkpoint, error) {
	cmd := exec.Command("git", "log", "--format=%H|%s|%at", fmt.Sprintf("-%d", limit))
	cmd.Dir = cs.workingDir
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var checkpoints []Checkpoint
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		hash := parts[0]
		message := parts[1]
		timestamp := time.Unix(parseUnixTimestamp(parts[2]), 0)

		checkpoints = append(checkpoints, Checkpoint{
			ID:        hash[:8], // Short hash
			Message:   message,
			Timestamp: timestamp,
			Hash:      hash,
			IsStashed: false,
		})
	}

	return checkpoints, nil
}

// restoreFromStash restores from a stash
func (cs *CheckpointService) restoreFromStash(checkpointID string) error {
	stashes, err := cs.getStashes()
	if err != nil {
		return fmt.Errorf("failed to get stashes: %w", err)
	}

	for i, stash := range stashes {
		if stash.ID == checkpointID {
			if err := cs.runGitCommand("stash", "apply", fmt.Sprintf("stash@{%d}", i)); err != nil {
				return fmt.Errorf("failed to apply stash: %w", err)
			}
			slog.Info("Restored checkpoint from stash", "id", checkpointID)
			return nil
		}
	}

	return fmt.Errorf("checkpoint not found: %s", checkpointID)
}

// restoreFromCommit restores from a commit (reset --hard)
func (cs *CheckpointService) restoreFromCommit(checkpointID string) error {
	if err := cs.runGitCommand("reset", "--hard", checkpointID); err != nil {
		return fmt.Errorf("failed to reset to commit: %w", err)
	}
	slog.Info("Restored checkpoint from commit", "id", checkpointID)
	return nil
}

// parseUnixTimestamp parses a unix timestamp string
func parseUnixTimestamp(s string) int64 {
	if ts := strings.TrimSpace(s); ts != "" {
		if t, err := time.Parse("1136239445", ts); err == nil {
			return t.Unix()
		}
	}
	return time.Now().Unix()
}

// getContextValues extracts session and message IDs from context
func getContextValues(ctx context.Context) (string, string) {
	sessionID, _ := ctx.Value("sessionID").(string)
	messageID, _ := ctx.Value("messageID").(string)
	return sessionID, messageID
}
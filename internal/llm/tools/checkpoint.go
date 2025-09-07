package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/crush/internal/checkpoint"
	"github.com/charmbracelet/crush/internal/permission"
)

type CheckpointParams struct {
	Action  string `json:"action"` // "create", "list", "restore", "delete"
	Message string `json:"message,omitempty"`
	ID      string `json:"id,omitempty"`
}

type checkpointTool struct {
	permissions       permission.Service
	checkpointService *checkpoint.CheckpointService
	workingDir        string
}

const CheckpointToolName = "checkpoint"

func NewCheckpointTool(permissions permission.Service, workingDir string) BaseTool {
	return &checkpointTool{
		permissions:       permissions,
		checkpointService: checkpoint.NewCheckpointService(workingDir, permissions),
		workingDir:        workingDir,
	}
}

func (t *checkpointTool) Info() ToolInfo {
	return ToolInfo{
		Name:        CheckpointToolName,
		Description: "Manage Git-based checkpoints to save and restore project state. Create checkpoints before making changes, list available checkpoints, and restore to previous states.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"enum":        []string{"create", "list", "restore", "delete"},
					"description": "Action to perform: create a new checkpoint, list checkpoints, restore to a checkpoint, or delete a checkpoint",
				},
				"message": map[string]any{
					"type":        "string",
					"description": "Description message for the checkpoint (required for create action)",
				},
				"id": map[string]any{
					"type":        "string",
					"description": "Checkpoint ID (required for restore and delete actions)",
				},
			},
			"required": []string{"action"},
		},
	}
}

func (t *checkpointTool) Name() string {
	return CheckpointToolName
}

func (t *checkpointTool) Run(ctx context.Context, params ToolCall) (ToolResponse, error) {
	var checkpointParams CheckpointParams
	if err := json.Unmarshal([]byte(params.Input), &checkpointParams); err != nil {
		return NewTextErrorResponse("Invalid parameters"), nil
	}

	switch checkpointParams.Action {
	case "create":
		if checkpointParams.Message == "" {
			return NewTextErrorResponse("Message is required for creating checkpoints"), nil
		}
		return t.createCheckpoint(ctx, checkpointParams.Message)

	case "list":
		return t.listCheckpoints(ctx)

	case "restore":
		if checkpointParams.ID == "" {
			return NewTextErrorResponse("ID is required for restoring checkpoints"), nil
		}
		return t.restoreCheckpoint(ctx, params.ID, checkpointParams.ID)

	case "delete":
		if checkpointParams.ID == "" {
			return NewTextErrorResponse("ID is required for deleting checkpoints"), nil
		}
		return t.deleteCheckpoint(ctx, checkpointParams.ID)

	default:
		return NewTextErrorResponse("Invalid action. Must be one of: create, list, restore, delete"), nil
	}
}

func (t *checkpointTool) createCheckpoint(ctx context.Context, message string) (ToolResponse, error) {
	checkpoint, err := t.checkpointService.CreateCheckpoint(ctx, message)
	if err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Failed to create checkpoint: %v", err)), nil
	}

	result := map[string]interface{}{
		"action":     "create",
		"success":    true,
		"checkpoint": checkpoint,
		"message":    fmt.Sprintf("Created checkpoint '%s' (ID: %s)", checkpoint.Message, checkpoint.ID),
	}

	output, _ := json.Marshal(result)
	return NewTextResponse(string(output)), nil
}

func (t *checkpointTool) listCheckpoints(ctx context.Context) (ToolResponse, error) {
	checkpoints, err := t.checkpointService.ListCheckpoints(ctx)
	if err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Failed to list checkpoints: %v", err)), nil
	}

	result := map[string]interface{}{
		"action":      "list",
		"success":     true,
		"checkpoints": checkpoints.Checkpoints,
		"count":       len(checkpoints.Checkpoints),
	}

	output, _ := json.Marshal(result)
	return NewTextResponse(string(output)), nil
}

func (t *checkpointTool) restoreCheckpoint(ctx context.Context, toolCallID, id string) (ToolResponse, error) {
	err := t.checkpointService.RestoreCheckpoint(ctx, id)
	if err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Failed to restore checkpoint: %v", err)), nil
	}

	result := map[string]interface{}{
		"action":  "restore",
		"success": true,
		"id":      id,
		"message": fmt.Sprintf("Successfully restored checkpoint %s", id),
	}

	output, _ := json.Marshal(result)
	return NewTextResponse(string(output)), nil
}

func (t *checkpointTool) deleteCheckpoint(ctx context.Context, id string) (ToolResponse, error) {
	err := t.checkpointService.DeleteCheckpoint(ctx, id)
	if err != nil {
		return NewTextErrorResponse(fmt.Sprintf("Failed to delete checkpoint: %v", err)), nil
	}

	result := map[string]interface{}{
		"action":  "delete",
		"success": true,
		"id":      id,
		"message": fmt.Sprintf("Successfully deleted checkpoint %s", id),
	}

	output, _ := json.Marshal(result)
	return NewTextResponse(string(output)), nil
}
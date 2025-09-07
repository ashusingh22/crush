package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/crush/internal/notifications"
	"github.com/charmbracelet/crush/internal/permission"
)

type NotificationParams struct {
	Service  string            `json:"service"`  // "discord", "telegram", "both"
	Title    string            `json:"title"`
	Message  string            `json:"message"`
	Level    string            `json:"level,omitempty"` // "info", "warning", "error", "success"
	Metadata map[string]string `json:"metadata,omitempty"`
}

type notificationTool struct {
	permissions     permission.Service
	discordService  *notifications.DiscordService
	telegramService *notifications.TelegramService
}

const NotificationToolName = "notify"

func NewNotificationTool(permissions permission.Service, config *notifications.NotificationConfig) BaseTool {
	var discordService *notifications.DiscordService
	var telegramService *notifications.TelegramService

	if config != nil {
		if config.Discord.Enabled {
			discordService = notifications.NewDiscordService(config.Discord)
		}
		if config.Telegram.Enabled {
			telegramService = notifications.NewTelegramService(config.Telegram)
		}
	}

	return &notificationTool{
		permissions:     permissions,
		discordService:  discordService,
		telegramService: telegramService,
	}
}

func (t *notificationTool) Info() ToolInfo {
	return ToolInfo{
		Name:        NotificationToolName,
		Description: "Send notifications via Discord webhooks or Telegram bot. Useful for alerting about task completion, errors, or important events.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"service": map[string]any{
					"type":        "string",
					"enum":        []string{"discord", "telegram", "both"},
					"description": "Notification service to use",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Notification title",
				},
				"message": map[string]any{
					"type":        "string",
					"description": "Notification message content",
				},
				"level": map[string]any{
					"type":        "string",
					"enum":        []string{"info", "warning", "error", "success"},
					"description": "Notification level (optional, defaults to 'info')",
				},
				"metadata": map[string]any{
					"type":        "object",
					"description": "Additional metadata to include (optional)",
				},
			},
			"required": []string{"service", "title", "message"},
		},
	}
}

func (t *notificationTool) Name() string {
	return NotificationToolName
}

func (t *notificationTool) Run(ctx context.Context, params ToolCall) (ToolResponse, error) {
	var notifyParams NotificationParams
	if err := json.Unmarshal([]byte(params.Input), &notifyParams); err != nil {
		return NewTextErrorResponse("Invalid parameters"), nil
	}

	// Validate required parameters
	if notifyParams.Title == "" {
		return NewTextErrorResponse("Title is required"), nil
	}
	if notifyParams.Message == "" {
		return NewTextErrorResponse("Message is required"), nil
	}

	// Set default level
	level := notifications.LevelInfo
	if notifyParams.Level != "" {
		switch notifyParams.Level {
		case "warning":
			level = notifications.LevelWarning
		case "error":
			level = notifications.LevelError
		case "success":
			level = notifications.LevelSuccess
		case "info":
			level = notifications.LevelInfo
		default:
			return NewTextErrorResponse("Invalid level. Must be one of: info, warning, error, success"), nil
		}
	}

	// Create notification
	notification := &notifications.Notification{
		Title:     notifyParams.Title,
		Message:   notifyParams.Message,
		Level:     level,
		Timestamp: time.Now(),
		Metadata:  notifyParams.Metadata,
	}

	// Check which services are available and requested
	var results []map[string]interface{}
	var errors []string

	if notifyParams.Service == "discord" || notifyParams.Service == "both" {
		if t.discordService != nil && t.discordService.IsEnabled() {
			if err := t.discordService.SendNotification(ctx, notification); err != nil {
				errors = append(errors, fmt.Sprintf("Discord: %v", err))
			} else {
				results = append(results, map[string]interface{}{
					"service": "discord",
					"success": true,
					"message": "Notification sent successfully",
				})
			}
		} else {
			errors = append(errors, "Discord service is not enabled or configured")
		}
	}

	if notifyParams.Service == "telegram" || notifyParams.Service == "both" {
		if t.telegramService != nil && t.telegramService.IsEnabled() {
			if err := t.telegramService.SendNotification(ctx, notification); err != nil {
				errors = append(errors, fmt.Sprintf("Telegram: %v", err))
			} else {
				results = append(results, map[string]interface{}{
					"service": "telegram",
					"success": true,
					"message": "Notification sent successfully",
				})
			}
		} else {
			errors = append(errors, "Telegram service is not enabled or configured")
		}
	}

	// Prepare response
	response := map[string]interface{}{
		"success":      len(errors) == 0,
		"results":      results,
		"notification": notification,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	if len(results) == 0 {
		return NewTextErrorResponse("No notifications were sent. Check service configuration."), nil
	}

	output, _ := json.Marshal(response)
	return NewTextResponse(string(output)), nil
}
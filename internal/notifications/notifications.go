package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// NotificationLevel represents the severity of a notification
type NotificationLevel string

const (
	LevelInfo    NotificationLevel = "info"
	LevelWarning NotificationLevel = "warning"
	LevelError   NotificationLevel = "error"
	LevelSuccess NotificationLevel = "success"
)

// Notification represents a notification to be sent
type Notification struct {
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Level     NotificationLevel `json:"level"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NotificationService defines the interface for notification services
type NotificationService interface {
	SendNotification(ctx context.Context, notification *Notification) error
	IsEnabled() bool
}

// DiscordConfig holds Discord webhook configuration
type DiscordConfig struct {
	WebhookURL string `json:"webhook_url"`
	Username   string `json:"username,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	Enabled    bool   `json:"enabled"`
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken  string `json:"bot_token"`
	ChatID    string `json:"chat_id"`
	Enabled   bool   `json:"enabled"`
}

// NotificationConfig holds all notification configurations
type NotificationConfig struct {
	Discord  DiscordConfig  `json:"discord,omitempty"`
	Telegram TelegramConfig `json:"telegram,omitempty"`
}

// DiscordService implements Discord notifications
type DiscordService struct {
	config DiscordConfig
	client *http.Client
}

// TelegramService implements Telegram notifications
type TelegramService struct {
	config TelegramConfig
	client *http.Client
}

// NewDiscordService creates a new Discord notification service
func NewDiscordService(config DiscordConfig) *DiscordService {
	return &DiscordService{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewTelegramService creates a new Telegram notification service
func NewTelegramService(config TelegramConfig) *TelegramService {
	return &TelegramService{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// IsEnabled returns whether Discord notifications are enabled
func (d *DiscordService) IsEnabled() bool {
	return d.config.Enabled && d.config.WebhookURL != ""
}

// IsEnabled returns whether Telegram notifications are enabled
func (t *TelegramService) IsEnabled() bool {
	return t.config.Enabled && t.config.BotToken != "" && t.config.ChatID != ""
}

// SendNotification sends a notification via Discord webhook
func (d *DiscordService) SendNotification(ctx context.Context, notification *Notification) error {
	if !d.IsEnabled() {
		return fmt.Errorf("Discord notifications are not enabled")
	}

	embed := map[string]interface{}{
		"title":       notification.Title,
		"description": notification.Message,
		"timestamp":   notification.Timestamp.Format(time.RFC3339),
		"color":       d.getColorForLevel(notification.Level),
	}

	if len(notification.Metadata) > 0 {
		fields := make([]map[string]interface{}, 0, len(notification.Metadata))
		for key, value := range notification.Metadata {
			fields = append(fields, map[string]interface{}{
				"name":   key,
				"value":  value,
				"inline": true,
			})
		}
		embed["fields"] = fields
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	if d.config.Username != "" {
		payload["username"] = d.config.Username
	}
	if d.config.AvatarURL != "" {
		payload["avatar_url"] = d.config.AvatarURL
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Discord request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Discord notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Discord API returned status %d", resp.StatusCode)
	}

	slog.Debug("Discord notification sent successfully", 
		"title", notification.Title, 
		"level", notification.Level)
	return nil
}

// SendNotification sends a notification via Telegram bot
func (t *TelegramService) SendNotification(ctx context.Context, notification *Notification) error {
	if !t.IsEnabled() {
		return fmt.Errorf("Telegram notifications are not enabled")
	}

	// Format message with emoji based on level
	emoji := t.getEmojiForLevel(notification.Level)
	message := fmt.Sprintf("%s *%s*\n\n%s", emoji, notification.Title, notification.Message)

	if len(notification.Metadata) > 0 {
		message += "\n\n*Details:*"
		for key, value := range notification.Metadata {
			message += fmt.Sprintf("\n• %s: %s", key, value)
		}
	}

	payload := map[string]interface{}{
		"chat_id":    t.config.ChatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Telegram payload: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.config.BotToken)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Telegram notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Telegram API returned status %d", resp.StatusCode)
	}

	slog.Debug("Telegram notification sent successfully", 
		"title", notification.Title, 
		"level", notification.Level)
	return nil
}

// getColorForLevel returns Discord embed color for notification level
func (d *DiscordService) getColorForLevel(level NotificationLevel) int {
	switch level {
	case LevelSuccess:
		return 0x00ff00 // Green
	case LevelWarning:
		return 0xffff00 // Yellow
	case LevelError:
		return 0xff0000 // Red
	default:
		return 0x0099ff // Blue
	}
}

// getEmojiForLevel returns emoji for notification level
func (t *TelegramService) getEmojiForLevel(level NotificationLevel) string {
	switch level {
	case LevelSuccess:
		return "✅"
	case LevelWarning:
		return "⚠️"
	case LevelError:
		return "❌"
	default:
		return "ℹ️"
	}
}
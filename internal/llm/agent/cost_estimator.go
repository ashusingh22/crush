package agent

import (
	"context"
	"log/slog"
	"strings"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/crush/internal/llm/provider"
	"github.com/charmbracelet/crush/internal/message"
)

// CostEstimator provides cost estimation for LLM requests
type CostEstimator struct {
	maxCostThreshold float64 // Maximum cost per request before warning
}

// NewCostEstimator creates a new cost estimator
func NewCostEstimator(maxCostThreshold float64) *CostEstimator {
	return &CostEstimator{
		maxCostThreshold: maxCostThreshold,
	}
}

// EstimateRequestCost estimates the cost of a request before making it
func (ce *CostEstimator) EstimateRequestCost(ctx context.Context, messages []message.Message, model catwalk.Model, maxTokens int) (*provider.TokenUsage, float64, error) {
	// Estimate input tokens
	inputTokens := ce.countTokensInMessages(messages)

	// Estimate output tokens (use maxTokens as upper bound, but use reasonable default)
	outputTokens := maxTokens
	if outputTokens == 0 {
		outputTokens = int(model.DefaultMaxTokens) / 2 // Conservative estimate
	}

	usage := provider.TokenUsage{
		InputTokens:  int64(inputTokens),
		OutputTokens: int64(outputTokens),
	}

	// Calculate cost
	cost := model.CostPer1MIn/1e6*float64(usage.InputTokens) +
		model.CostPer1MOut/1e6*float64(usage.OutputTokens)

	slog.Debug("Cost estimation",
		"input_tokens", usage.InputTokens,
		"output_tokens", usage.OutputTokens,
		"estimated_cost", cost,
		"model", model.ID,
	)

	return &usage, cost, nil
}

// ShouldProceed checks if the request should proceed based on cost
func (ce *CostEstimator) ShouldProceed(estimatedCost float64) (bool, string) {
	if estimatedCost > ce.maxCostThreshold {
		return false, "Estimated cost exceeds threshold"
	}
	return true, ""
}

// countTokensInMessages provides a rough token count estimate
// This is a simplified implementation - real token counting would use the model's tokenizer
func (ce *CostEstimator) countTokensInMessages(messages []message.Message) int {
	totalTokens := 0

	for _, msg := range messages {
		// Add base tokens for role and structure
		totalTokens += 4

		for _, part := range msg.Parts {
			switch p := part.(type) {
			case message.TextContent:
				totalTokens += ce.estimateTextTokens(p.Text)
			case message.ToolCall:
				totalTokens += ce.estimateTextTokens(p.Name)
				totalTokens += ce.estimateTextTokens(p.Input)
			case message.ToolResult:
				totalTokens += ce.estimateTextTokens(p.Content)
			}
		}
	}

	return totalTokens
}

// estimateTextTokens provides a rough estimate of tokens in text
func (ce *CostEstimator) estimateTextTokens(text string) int {
	// Rough approximation: 1 token per 4 characters for English text
	// This varies by model and language, but provides a reasonable estimate
	words := len(strings.Fields(text))
	chars := len(text)

	// Use a heuristic that combines word count and character count
	// This tends to be more accurate than just character count
	estimate := int(float64(words)*1.3 + float64(chars)*0.25)

	return estimate
}

// OptimizeMessages attempts to reduce message size while preserving important context
func (ce *CostEstimator) OptimizeMessages(ctx context.Context, messages []message.Message, targetReduction float64) []message.Message {
	if targetReduction <= 0 || targetReduction >= 1 {
		return messages
	}

	optimized := make([]message.Message, 0, len(messages))
	currentSize := ce.countTokensInMessages(messages)
	targetSize := int(float64(currentSize) * (1 - targetReduction))

	slog.Debug("Optimizing messages",
		"original_tokens", currentSize,
		"target_tokens", targetSize,
		"reduction", targetReduction,
	)

	// Keep the last few messages (most recent context) and system message
	importantCount := min(5, len(messages))

	// Always include system messages and recent messages
	for i, msg := range messages {
		if msg.Role == message.System || i >= len(messages)-importantCount {
			optimized = append(optimized, msg)
			continue
		}

		// For older messages, check if we need to truncate
		if ce.countTokensInMessages(optimized) < targetSize {
			// Try to summarize or truncate this message
			summarized := ce.summarizeMessage(msg)
			optimized = append(optimized, summarized)
		}
	}

	finalSize := ce.countTokensInMessages(optimized)
	slog.Debug("Message optimization complete",
		"original_tokens", currentSize,
		"final_tokens", finalSize,
		"reduction_achieved", float64(currentSize-finalSize)/float64(currentSize),
	)

	return optimized
}

// summarizeMessage creates a shorter version of a message while preserving key information
func (ce *CostEstimator) summarizeMessage(msg message.Message) message.Message {
	summarized := msg
	summarized.Parts = nil

	for _, part := range msg.Parts {
		switch p := part.(type) {
		case message.TextContent:
			// Truncate long text content
			content := p.Text
			if len(content) > 500 {
				content = content[:400] + "... [truncated]"
			}
			summarized.Parts = append(summarized.Parts, message.TextContent{Text: content})
		case message.ToolCall:
			// Keep tool calls as they're usually important
			summarized.Parts = append(summarized.Parts, p)
		case message.ToolResult:
			// Summarize tool results if they're long
			content := p.Content
			if len(content) > 1000 {
				content = content[:800] + "... [result truncated]"
			}
			summarized.Parts = append(summarized.Parts, message.ToolResult{
				ToolCallID: p.ToolCallID,
				Content:    content,
				IsError:    p.IsError,
			})
		default:
			summarized.Parts = append(summarized.Parts, part)
		}
	}

	return summarized
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

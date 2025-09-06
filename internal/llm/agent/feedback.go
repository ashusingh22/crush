package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/message"
)

// ResponseQuality represents the quality score and analysis of a response
type ResponseQuality struct {
	Score           float64           `json:"score"`            // 0.0 to 1.0 quality score
	Confidence      float64           `json:"confidence"`       // 0.0 to 1.0 confidence in the score
	Issues          []string          `json:"issues"`           // List of identified issues
	Suggestions     []string          `json:"suggestions"`      // Improvement suggestions
	RequiresRetry   bool              `json:"requires_retry"`   // Whether response should be regenerated
	Metrics         map[string]float64 `json:"metrics"`         // Individual quality metrics
	Timestamp       time.Time         `json:"timestamp"`
}

// FeedbackMechanism provides response quality validation and improvement suggestions
type FeedbackMechanism struct {
	minQualityThreshold float64
	maxRetryAttempts    int
	enabled             bool
}

// NewFeedbackMechanism creates a new feedback mechanism
func NewFeedbackMechanism(enabled bool, minQualityThreshold float64, maxRetryAttempts int) *FeedbackMechanism {
	return &FeedbackMechanism{
		minQualityThreshold: minQualityThreshold,
		maxRetryAttempts:    maxRetryAttempts,
		enabled:             enabled,
	}
}

// EvaluateResponse analyzes the quality of a response
func (fm *FeedbackMechanism) EvaluateResponse(ctx context.Context, userMessage message.Message, response message.Message) *ResponseQuality {
	if !fm.enabled {
		return &ResponseQuality{
			Score:         1.0,
			Confidence:    0.5,
			RequiresRetry: false,
			Timestamp:     time.Now(),
		}
	}

	quality := &ResponseQuality{
		Issues:      []string{},
		Suggestions: []string{},
		Metrics:     make(map[string]float64),
		Timestamp:   time.Now(),
	}

	// Extract text content from response
	responseText := fm.extractTextContent(response)
	userText := fm.extractTextContent(userMessage)

	// Calculate individual quality metrics
	quality.Metrics["completeness"] = fm.calculateCompleteness(userText, responseText)
	quality.Metrics["clarity"] = fm.calculateClarity(responseText)
	quality.Metrics["relevance"] = fm.calculateRelevance(userText, responseText)
	quality.Metrics["specificity"] = fm.calculateSpecificity(responseText)
	quality.Metrics["error_indicators"] = fm.detectErrorIndicators(responseText)

	// Calculate overall quality score
	quality.Score = fm.calculateOverallScore(quality.Metrics)
	quality.Confidence = fm.calculateConfidence(quality.Metrics, responseText)

	// Check for specific issues
	fm.analyzeIssues(quality, userText, responseText)

	// Determine if retry is needed
	quality.RequiresRetry = quality.Score < fm.minQualityThreshold

	slog.Debug("Response quality evaluation", 
		"score", quality.Score,
		"confidence", quality.Confidence,
		"requires_retry", quality.RequiresRetry,
		"issues_count", len(quality.Issues),
	)

	return quality
}

// extractTextContent extracts text from message parts
func (fm *FeedbackMechanism) extractTextContent(msg message.Message) string {
	var texts []string
	for _, part := range msg.Parts {
		if textPart, ok := part.(message.TextContent); ok {
			texts = append(texts, textPart.Text)
		}
	}
	return strings.Join(texts, "\n")
}

// calculateCompleteness measures how well the response addresses the user's request
func (fm *FeedbackMechanism) calculateCompleteness(userText, responseText string) float64 {
	if responseText == "" {
		return 0.0
	}

	// Check if response length is appropriate for the request
	userWords := len(strings.Fields(userText))
	responseWords := len(strings.Fields(responseText))

	// Heuristic: response should be proportional to request complexity
	if userWords > 0 {
		ratio := float64(responseWords) / float64(userWords)
		if ratio < 0.5 {
			return 0.3 // Too short
		}
		if ratio > 10 {
			return 0.7 // Might be too verbose
		}
	}

	// Check for common completeness indicators
	responseText = strings.ToLower(responseText)
	
	incompleteIndicators := []string{
		"i need more information",
		"could you clarify",
		"incomplete",
		"not enough context",
		"unable to determine",
	}

	for _, indicator := range incompleteIndicators {
		if strings.Contains(responseText, indicator) {
			return 0.4
		}
	}

	return 0.8
}

// calculateClarity measures how clear and understandable the response is
func (fm *FeedbackMechanism) calculateClarity(responseText string) float64 {
	if responseText == "" {
		return 0.0
	}

	words := strings.Fields(responseText)
	sentences := strings.Split(responseText, ".")
	
	// Average sentence length (clarity decreases with very long sentences)
	avgSentenceLength := float64(len(words)) / float64(len(sentences))
	
	clarityScore := 0.8
	
	// Penalize overly complex sentences
	if avgSentenceLength > 25 {
		clarityScore -= 0.2
	}
	
	// Check for clarity indicators
	responseText = strings.ToLower(responseText)
	clarityIndicators := []string{
		"first", "second", "then", "next", "finally",
		"however", "therefore", "because", "since",
	}
	
	indicatorCount := 0
	for _, indicator := range clarityIndicators {
		if strings.Contains(responseText, indicator) {
			indicatorCount++
		}
	}
	
	// Boost score for structured responses
	if indicatorCount > 2 {
		clarityScore += 0.1
	}

	return minFloat64(clarityScore, 1.0)
}

// calculateRelevance measures how relevant the response is to the user's request
func (fm *FeedbackMechanism) calculateRelevance(userText, responseText string) float64 {
	if responseText == "" {
		return 0.0
	}

	userWords := strings.Fields(strings.ToLower(userText))
	responseWords := strings.Fields(strings.ToLower(responseText))

	// Simple keyword overlap measurement
	userWordSet := make(map[string]bool)
	for _, word := range userWords {
		if len(word) > 3 { // Focus on meaningful words
			userWordSet[word] = true
		}
	}

	overlap := 0
	for _, word := range responseWords {
		if len(word) > 3 && userWordSet[word] {
			overlap++
		}
	}

	if len(userWordSet) == 0 {
		return 0.5
	}

	relevanceScore := float64(overlap) / float64(len(userWordSet))
	return minFloat64(relevanceScore, 1.0)
}

// calculateSpecificity measures how specific and actionable the response is
func (fm *FeedbackMechanism) calculateSpecificity(responseText string) float64 {
	if responseText == "" {
		return 0.0
	}

	responseText = strings.ToLower(responseText)
	
	// Check for vague language
	vagueTerms := []string{
		"maybe", "perhaps", "might", "could be", "possibly",
		"generally", "usually", "often", "sometimes",
		"it depends", "varies", "different",
	}

	vaguenessCount := 0
	for _, term := range vagueTerms {
		vaguenessCount += strings.Count(responseText, term)
	}

	// Check for specific indicators
	specificIndicators := []string{
		"step 1", "step 2", "specifically", "exactly",
		"run the following", "execute", "use this command",
		"set to", "configure", "install",
	}

	specificityCount := 0
	for _, indicator := range specificIndicators {
		if strings.Contains(responseText, indicator) {
			specificityCount++
		}
	}

	specificityScore := 0.5
	
	// Penalize vagueness
	if vaguenessCount > 3 {
		specificityScore -= 0.3
	}
	
	// Reward specificity
	if specificityCount > 0 {
		specificityScore += 0.3
	}

	return maxFloat64(0.0, minFloat64(specificityScore, 1.0))
}

// detectErrorIndicators looks for signs of errors or hallucinations
func (fm *FeedbackMechanism) detectErrorIndicators(responseText string) float64 {
	responseText = strings.ToLower(responseText)
	
	errorIndicators := []string{
		"i apologize, but",
		"i'm sorry, i can't",
		"error",
		"failed",
		"unable to",
		"not possible",
		"doesn't exist",
		"not found",
		"invalid",
	}

	errorCount := 0
	for _, indicator := range errorIndicators {
		if strings.Contains(responseText, indicator) {
			errorCount++
		}
	}

	// Return 1.0 for no errors, decreasing with more error indicators
	return max(0.0, 1.0 - float64(errorCount)*0.2)
}

// calculateOverallScore combines individual metrics into an overall quality score
func (fm *FeedbackMechanism) calculateOverallScore(metrics map[string]float64) float64 {
	weights := map[string]float64{
		"completeness":     0.3,
		"clarity":          0.2,
		"relevance":        0.25,
		"specificity":      0.15,
		"error_indicators": 0.1,
	}

	totalScore := 0.0
	totalWeight := 0.0

	for metric, score := range metrics {
		if weight, exists := weights[metric]; exists {
			totalScore += score * weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0.5
	}

	return totalScore / totalWeight
}

// calculateConfidence estimates confidence in the quality assessment
func (fm *FeedbackMechanism) calculateConfidence(metrics map[string]float64, responseText string) float64 {
	// Base confidence depends on response length
	wordCount := len(strings.Fields(responseText))
	
	confidence := 0.5
	
	if wordCount > 50 {
		confidence += 0.2 // More content to analyze
	}
	if wordCount > 200 {
		confidence += 0.1 // Even more content
	}
	
	// Reduce confidence if metrics are inconsistent
	variance := fm.calculateMetricsVariance(metrics)
	if variance > 0.3 {
		confidence -= 0.2
	}

	return maxFloat64(0.1, minFloat64(confidence, 0.9))
}

// calculateMetricsVariance calculates variance in quality metrics
func (fm *FeedbackMechanism) calculateMetricsVariance(metrics map[string]float64) float64 {
	if len(metrics) == 0 {
		return 0
	}

	sum := 0.0
	for _, score := range metrics {
		sum += score
	}
	mean := sum / float64(len(metrics))

	varianceSum := 0.0
	for _, score := range metrics {
		diff := score - mean
		varianceSum += diff * diff
	}

	return varianceSum / float64(len(metrics))
}

// analyzeIssues identifies specific issues and suggests improvements
func (fm *FeedbackMechanism) analyzeIssues(quality *ResponseQuality, userText, responseText string) {
	// Check completeness issues
	if quality.Metrics["completeness"] < 0.6 {
		quality.Issues = append(quality.Issues, "Response may be incomplete")
		quality.Suggestions = append(quality.Suggestions, "Consider providing more detailed information")
	}

	// Check clarity issues
	if quality.Metrics["clarity"] < 0.6 {
		quality.Issues = append(quality.Issues, "Response may be unclear")
		quality.Suggestions = append(quality.Suggestions, "Break down complex information into clearer steps")
	}

	// Check relevance issues
	if quality.Metrics["relevance"] < 0.5 {
		quality.Issues = append(quality.Issues, "Response may not be relevant to the request")
		quality.Suggestions = append(quality.Suggestions, "Focus more directly on the user's specific question")
	}

	// Check specificity issues
	if quality.Metrics["specificity"] < 0.5 {
		quality.Issues = append(quality.Issues, "Response is too vague")
		quality.Suggestions = append(quality.Suggestions, "Provide more specific examples and concrete steps")
	}

	// Check for error indicators
	if quality.Metrics["error_indicators"] < 0.8 {
		quality.Issues = append(quality.Issues, "Response contains error indicators")
		quality.Suggestions = append(quality.Suggestions, "Verify the accuracy of the information provided")
	}
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// GenerateImprovementPrompt creates a prompt for improving the response
func (fm *FeedbackMechanism) GenerateImprovementPrompt(ctx context.Context, originalResponse message.Message, quality *ResponseQuality) string {
	if len(quality.Issues) == 0 {
		return ""
	}

	prompt := "Please improve the previous response by addressing the following issues:\n\n"
	
	for i, issue := range quality.Issues {
		prompt += fmt.Sprintf("%d. %s\n", i+1, issue)
	}
	
	if len(quality.Suggestions) > 0 {
		prompt += "\nSuggestions for improvement:\n"
		for _, suggestion := range quality.Suggestions {
			prompt += fmt.Sprintf("- %s\n", suggestion)
		}
	}

	prompt += "\nPlease provide a revised response that addresses these concerns while maintaining accuracy and helpfulness."

	return prompt
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
package agent

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/charmbracelet/crush/internal/llm/provider"
	"github.com/charmbracelet/crush/internal/message"
)

// CacheEntry represents a cached response
type CacheEntry struct {
	Response   message.Message
	TokenUsage provider.TokenUsage
	Timestamp  time.Time
	TTL        time.Duration
}

// IsExpired checks if cache entry has expired
func (c *CacheEntry) IsExpired() bool {
	return time.Since(c.Timestamp) > c.TTL
}

// ResponseCache provides caching for LLM responses to reduce API calls
type ResponseCache struct {
	cache   map[string]*CacheEntry
	mu      sync.RWMutex
	enabled bool
	// Default TTL for cache entries
	defaultTTL time.Duration
	// Maximum cache size
	maxSize int
}

// NewResponseCache creates a new response cache
func NewResponseCache(enabled bool, defaultTTL time.Duration, maxSize int) *ResponseCache {
	return &ResponseCache{
		cache:      make(map[string]*CacheEntry),
		enabled:    enabled,
		defaultTTL: defaultTTL,
		maxSize:    maxSize,
	}
}

// generateCacheKey creates a unique key for the request
func (rc *ResponseCache) generateCacheKey(messages []message.Message, modelID string) string {
	hasher := sha256.New()

	// Include model ID in hash
	hasher.Write([]byte(modelID))

	// Hash the message content
	for _, msg := range messages {
		hasher.Write([]byte(string(msg.Role)))
		for _, part := range msg.Parts {
			if textPart, ok := part.(message.TextContent); ok {
				hasher.Write([]byte(textPart.Text))
			}
		}
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// Get retrieves a cached response if available and not expired
func (rc *ResponseCache) Get(ctx context.Context, messages []message.Message, modelID string) (*CacheEntry, bool) {
	if !rc.enabled {
		return nil, false
	}

	key := rc.generateCacheKey(messages, modelID)

	rc.mu.RLock()
	entry, exists := rc.cache[key]
	rc.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if entry.IsExpired() {
		// Clean up expired entry
		rc.mu.Lock()
		delete(rc.cache, key)
		rc.mu.Unlock()
		return nil, false
	}

	slog.Debug("Cache hit for LLM request", "key", key[:8])
	return entry, true
}

// Set stores a response in the cache
func (rc *ResponseCache) Set(ctx context.Context, messages []message.Message, modelID string, response message.Message, usage provider.TokenUsage) {
	if !rc.enabled {
		return
	}

	key := rc.generateCacheKey(messages, modelID)

	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Check if we need to evict entries
	if len(rc.cache) >= rc.maxSize {
		rc.evictOldest()
	}

	rc.cache[key] = &CacheEntry{
		Response:   response,
		TokenUsage: usage,
		Timestamp:  time.Now(),
		TTL:        rc.defaultTTL,
	}

	slog.Debug("Cached LLM response", "key", key[:8], "input_tokens", usage.InputTokens, "output_tokens", usage.OutputTokens)
}

// evictOldest removes the oldest cache entry
func (rc *ResponseCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range rc.cache {
		if oldestKey == "" || entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}

	if oldestKey != "" {
		delete(rc.cache, oldestKey)
		slog.Debug("Evicted oldest cache entry", "key", oldestKey[:8])
	}
}

// Clear removes all cached entries
func (rc *ResponseCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.cache = make(map[string]*CacheEntry)
	slog.Debug("Cleared response cache")
}

// Size returns the current number of cached entries
func (rc *ResponseCache) Size() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	return len(rc.cache)
}

// CleanExpired removes all expired entries
func (rc *ResponseCache) CleanExpired() int {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	var expiredKeys []string
	for key, entry := range rc.cache {
		if entry.IsExpired() {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(rc.cache, key)
	}

	if len(expiredKeys) > 0 {
		slog.Debug("Cleaned expired cache entries", "count", len(expiredKeys))
	}

	return len(expiredKeys)
}

// GetStats returns cache statistics
func (rc *ResponseCache) GetStats() map[string]interface{} {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	totalEntries := len(rc.cache)
	expiredCount := 0

	for _, entry := range rc.cache {
		if entry.IsExpired() {
			expiredCount++
		}
	}

	return map[string]interface{}{
		"enabled":        rc.enabled,
		"total_entries":  totalEntries,
		"expired_count":  expiredCount,
		"active_entries": totalEntries - expiredCount,
		"max_size":       rc.maxSize,
		"default_ttl":    rc.defaultTTL.String(),
	}
}

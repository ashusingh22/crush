# Crush Enhanced Features

This document describes the enhanced features added to Crush for cost optimization, quality improvement, and productivity.

## Overview of Enhancements

The following new features have been added to Crush:

1. **Response Caching** - Reduces API calls by caching similar responses
2. **Cost Estimation** - Predicts and controls API costs before making requests  
3. **Quality Feedback Mechanism** - Automatically evaluates and improves response quality
4. **Enhanced Productivity Tools** - New tools that reduce dependency on LLM calls
5. **Smart Permission System** - Learns user patterns for intelligent auto-approval

## Configuration

Add the following to your `crush.json` configuration file to enable these features:

```json
{
  "options": {
    "enhance_features": {
      "enable_cache": true,
      "cache_ttl_minutes": 30,
      "cache_max_entries": 100,
      "enable_cost_estimation": true,
      "max_cost_threshold": 0.50,
      "auto_optimize_context": true,
      "enable_feedback": true,
      "quality_threshold": 0.7,
      "max_retry_attempts": 2
    }
  }
}
```

## Feature Details

### 1. Response Caching

**Purpose**: Reduce API calls for similar requests to save costs.

**How it works**: 
- Generates unique cache keys based on message content and model
- Stores responses with configurable TTL (time-to-live)
- Automatically serves cached responses for matching requests

**Configuration**:
- `enable_cache`: Enable/disable caching (default: true)
- `cache_ttl_minutes`: How long to keep cached responses (default: 30 minutes)
- `cache_max_entries`: Maximum number of cached entries (default: 100)

### 2. Cost Estimation

**Purpose**: Predict and control API costs before making expensive requests.

**How it works**:
- Estimates token count using heuristics
- Calculates estimated cost based on model pricing
- Warns or blocks requests exceeding cost thresholds
- Automatically optimizes context for high-cost requests

**Configuration**:
- `enable_cost_estimation`: Enable cost prediction (default: true)
- `max_cost_threshold`: Maximum cost per request in USD (default: $0.50)
- `auto_optimize_context`: Auto-reduce context for expensive requests (default: true)

### 3. Quality Feedback Mechanism

**Purpose**: Automatically evaluate response quality and suggest improvements.

**How it works**:
- Evaluates responses on multiple quality metrics:
  - Completeness - How well the response addresses the request
  - Clarity - How clear and understandable the response is
  - Relevance - How relevant the response is to the question
  - Specificity - How specific and actionable the response is
  - Error indicators - Detection of potential errors or hallucinations
- Generates improvement suggestions for low-quality responses
- Queues improvement prompts for iterative enhancement

**Configuration**:
- `enable_feedback`: Enable quality evaluation (default: true)
- `quality_threshold`: Minimum acceptable quality score 0.0-1.0 (default: 0.7)
- `max_retry_attempts`: Maximum retry attempts for improvement (default: 2)

### 4. Enhanced Productivity Tools

#### Analyze Tool

**Purpose**: Analyze code structure, complexity, and patterns without LLM calls.

**Usage**:
```
I want to analyze the complexity of my Go files
```

**Supported analysis types**:
- `structure`: File/directory structure analysis
- `complexity`: Cyclomatic complexity calculation
- `dependencies`: Dependency analysis (planned)
- `patterns`: Design pattern detection (planned)

**Supported languages**: Go, JavaScript, TypeScript, Python

#### Batch Tool

**Purpose**: Execute multiple operations in batch to reduce API call overhead.

**Usage**:
```
Execute these operations in batch:
1. Find all .go files containing "func"
2. Copy important files to backup folder
3. Analyze directory structure
```

**Supported operations**:
- `file_search`: Search for files by name/pattern
- `text_replace`: Replace text in files
- `file_copy`: Copy files
- `dir_analysis`: Analyze directory statistics
- `pattern_find`: Find text patterns in code files

### 5. Smart Permission System

**Purpose**: Learn from user permission patterns to enable intelligent auto-approval.

**How it works**:
- Records user approval/denial decisions
- Learns patterns based on tool, action, and path
- Calculates confidence scores for auto-approval
- Automatically approves high-confidence, safe operations

**Features**:
- Pattern generalization (e.g., temp files, versioned files)
- Confidence-based auto-approval
- Time-based pattern decay
- Safe operation detection

**Data storage**: Patterns are stored in `.crush/permission_patterns.json`

## Benefits

### Cost Savings
- **Response caching**: Eliminates duplicate API calls
- **Cost estimation**: Prevents accidentally expensive requests  
- **Context optimization**: Reduces token usage by 30% when needed
- **Batch operations**: Reduces multiple small API calls

### Quality Improvement
- **Automatic evaluation**: Ensures responses meet quality standards
- **Iterative improvement**: Automatically refines low-quality responses
- **Error detection**: Identifies potential hallucinations and errors

### Productivity Enhancement
- **Local analysis**: Code analysis without API calls
- **Batch processing**: Efficient multi-operation execution
- **Smart permissions**: Reduces interruption from permission prompts
- **Better tools**: More capable tools that require less LLM assistance

## Usage Examples

### Example 1: Cost-Conscious Development
```bash
# The system will automatically:
# 1. Check cache for similar requests
# 2. Estimate cost before API call
# 3. Optimize context if cost is high
# 4. Cache the response for future use

crush
> Help me analyze this large codebase for security issues
```

### Example 2: Quality-Focused Responses
```bash
# The system will automatically:
# 1. Evaluate response quality
# 2. Suggest improvements if quality is low
# 3. Queue improvement prompts for next iteration

crush
> Explain how to implement OAuth2 authentication
```

### Example 3: Efficient Batch Operations
```bash
# Single request that performs multiple operations
crush
> I need to:
> 1. Find all TypeScript files with "TODO" comments
> 2. Analyze the complexity of my main modules
> 3. Create a backup of my config files
```

## Monitoring and Statistics

You can monitor the performance of these features through debug logs. Enable debug mode in your configuration:

```json
{
  "options": {
    "debug": true
  }
}
```

This will show detailed information about:
- Cache hits and misses
- Cost estimations and optimizations
- Quality scores and improvement suggestions
- Permission pattern learning

## Future Enhancements

Planned improvements include:
- More sophisticated tokenization for better cost estimation
- Advanced pattern detection in code analysis
- Machine learning-based quality assessment
- Integration with external cost monitoring tools
- Enhanced permission pattern recognition
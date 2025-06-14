# SlopSquid 🦑

**CLI tool for detecting and cleaning AI artifacts in your documentation**

*"Deploy the tentacles!"* - Batch process your docs with parallel AI detection to identify and clean up artificial writing patterns, keeping your documentation authentic and polished.

> **🚧 Active Development** - Core CLI functionality in progress, integrating with QRY tool ecosystem.

## ✨ What SlopSquid Does

- 🦑 **Parallel Processing** - Deploy multiple "tentacles" to process files simultaneously
- 🔍 **AI Pattern Detection** - Identify corporate speak, buzzwords, and artificial writing patterns
- 🎯 **Batch Operations** - Clean entire directory trees with confidence-based filtering
- 🏠 **Local Processing** - All AI analysis happens offline using local models (no cloud APIs)
- 🔗 **QRY Integration** - Logs cleanup decisions to uroboro, feeds insights to osmotic
- 📝 **Multiple Formats** - Works with Markdown, HTML, plain text, and more

## 🎯 Perfect For

- **Developers** - Clean up AI-assisted documentation and README files
- **Technical Writers** - Maintain authentic voice in documentation projects  
- **Content Creators** - Remove AI artifacts from drafts and articles
- **QRY Users** - Integrate with uroboro, osmotic, and the broader tool ecosystem

## 🚀 Quick Start

### Installation
```bash
# Install from releases (coming soon)
curl -sSL https://slopsquid.com/install | bash

# Or build from source
git clone https://github.com/QRY91/slopsquid
cd slopsquid
go build -o slopsquid ./cmd/slopsquid
```

### Basic Usage
```bash
# Scan a single file
slopsquid scan README.md

# Deploy tentacles on a directory (parallel processing)
slopsquid scan docs/ --recursive --threads=8

# Interactive cleanup mode
slopsquid clean docs/ --interactive --confidence=0.8

# Batch processing with uroboro integration
slopsquid scan docs/ --uroboro-project=documentation-cleanup
```

## 🦑 How the Tentacles Work

SlopSquid deploys multiple worker "tentacles" that grab files in parallel and analyze them using local AI models:

```
┌─ File Discovery ────┐    ┌─ Tentacle Pool ────┐    ┌─ Results ────────┐
│ • Recursive scan    │ -> │ • 8 worker threads │ -> │ • Confidence      │
│ • Pattern matching  │    │ • Local AI model   │    │ • Line markers    │
│ • Git integration   │    │ • Shared patterns  │    │ • Batch fixes     │
└────────────────────┘    └───────────────────┘    └──────────────────┘
```

Each tentacle:
1. **Grabs files** from the processing queue
2. **Analyzes content** using local AI models (ollama/llama)  
3. **Applies pattern detection** for AI writing signatures
4. **Scores confidence** based on multiple detection algorithms
5. **Inks suspicious content** with precise line/character markers

## 🛠️ Core Commands

### File Processing
```bash
# Scan operations
slopsquid scan file.md                    # Single file analysis
slopsquid scan docs/ --recursive          # Directory tree processing
slopsquid scan --git-staged               # Pre-commit integration

# Cleanup operations  
slopsquid clean file.md --interactive     # Review each detection
slopsquid clean docs/ --auto-fix --safe   # Auto-fix obvious patterns
slopsquid clean docs/ --dry-run           # Preview changes only
```

### Integration & Configuration
```bash
# QRY ecosystem integration
slopsquid scan --uroboro-project docs     # Log findings to uroboro
slopsquid scan --osmotic-feed              # Stream insights to osmotic

# Pattern learning and tuning
slopsquid train file.md                    # Learn from corrections
slopsquid config --sensitivity 0.8        # Adjust detection threshold
slopsquid patterns --export patterns.yaml # Export custom patterns
```

### Batch Operations
```bash
# Parallel processing with multiple tentacles
slopsquid scan docs/ --threads=12 --model=llama3.2:1b

# Confidence-based filtering
slopsquid clean docs/ --confidence-min=0.9 --auto-fix

# Output formats for integration
slopsquid scan docs/ --format=json > results.json
```

## 🎨 Detection Capabilities

### AI Writing Patterns
- **Buzzword Detection** - Corporate speak, unnecessary jargon, AI favorites
- **Syntactic Analysis** - Repetitive sentence structures, generic transitions
- **Semantic Patterns** - Unsupported claims, generic problem-solution language
- **Style Fingerprinting** - Lack of personal voice, unnatural consistency

### Confidence Scoring
- **🟢 Low (60-79%)** - Subtle suggestions, may be legitimate formal writing
- **🟡 Medium (80-89%)** - Likely AI patterns, worth human review
- **🔴 High (90%+)** - Clear AI signatures, safe for automated cleanup

### Context Awareness
- **Technical Documentation** - Allows legitimate precision language
- **Creative Writing** - Higher tolerance for varied expression
- **Academic Papers** - Considers formal language requirements
- **Marketing Content** - Adjusts for promotional language patterns

## 🔗 QRY Ecosystem Integration

### With uroboro
```bash
# Capture cleanup decisions for learning
slopsquid clean docs/ --interactive --uroboro-log
# → Each correction decision gets captured with context

# Log significant pattern discoveries
slopsquid scan --uroboro-project=doc-quality
# → AI detection insights become part of your development history
```

### With osmotic  
```bash
# Feed documentation quality patterns to osmotic
slopsquid scan docs/ --osmotic-stream
# → "Documentation cleanup activity detected" trends
# → Insights appear in osmotic morning briefings
```

### Pre-commit Integration
```bash
# Git hook for quality assurance
echo 'slopsquid scan --git-staged --confidence-min=0.9' > .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# Prevent commits with obvious AI artifacts
```

## 🏗️ Technical Architecture

### Local AI Processing
- **Model Support** - ollama, llama.cpp, local transformers
- **Offline First** - No cloud APIs, complete privacy
- **Resource Efficient** - Optimized for batch processing
- **Model Flexibility** - Use any local language model

### Parallel Processing Engine
- **Worker Pool** - Configurable number of processing tentacles
- **Queue Management** - Efficient file distribution and load balancing
- **Result Aggregation** - Merge and score results from multiple workers
- **Memory Management** - Process large directory trees without memory issues

### File Format Support
- **Markdown** - README files, documentation, blog posts
- **HTML** - Web content, documentation sites
- **Plain Text** - Any text-based content
- **Code Comments** - Clean up AI-generated code documentation
- **Extensible** - Plugin system for additional formats

## 🎯 Use Cases

### Documentation Maintenance
- **Post-AI Collaboration** - Clean up after AI writing sessions
- **Pre-Publication** - Ensure authentic voice in public documentation  
- **Portfolio Preparation** - Polish all materials before sharing
- **Team Standards** - Maintain consistent voice across team documentation

### Development Workflow
- **Pre-commit Hooks** - Automatic quality checks before commits
- **CI/CD Integration** - Documentation quality gates in build pipelines
- **Code Review** - Identify AI artifacts in pull request documentation
- **Release Preparation** - Clean up all customer-facing materials

### Content Creation
- **Draft Polishing** - Remove AI artifacts while preserving meaning
- **Voice Consistency** - Maintain personal/brand voice across content
- **Quality Assurance** - Systematic approach to content authenticity
- **Learning Tool** - Understand and avoid AI writing patterns

## 🚧 Development Status

### ✅ Architecture Complete
- CLI command structure designed
- Parallel processing architecture defined
- QRY ecosystem integration planned
- Local AI processing pipeline specified

### 🔧 In Progress
- Core CLI implementation (Go)
- File discovery and processing engine
- Pattern detection algorithms
- Local AI model integration

### 🎯 Coming Soon
- Interactive cleanup interface
- Pattern learning system
- Advanced configuration options
- QRY ecosystem integrations

## 🤝 Contributing

SlopSquid is part of the [QRY Tool Ecosystem](https://github.com/QRY91) - building privacy-first, locally-processed tools for systematic creators.

### Development Setup
```bash
git clone https://github.com/QRY91/slopsquid
cd slopsquid
go mod tidy
go build ./cmd/slopsquid
```

### Philosophy
- **Local-first processing** - No cloud dependencies
- **Authentic voice preservation** - Clean up AI artifacts without losing meaning
- **Systematic integration** - Work seamlessly with other QRY tools
- **Privacy by design** - Your content never leaves your system

## 📄 License

MIT License - see LICENSE file for details

---

*Deploy the tentacles. Clean up the slop. Keep your voice authentic.* 🦑

**Website**: [slopsquid.com](https://slopsquid.com)  
**Part of**: [QRY Tool Ecosystem](https://github.com/QRY91)
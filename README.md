# Converso CLI

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)

**Converso CLI** - Enterprise SaaS Command Line Interface

A production-ready, enterprise-grade CLI tool built with Go and Python, designed for Converso Empire's unified platform access.

## 🚀 Features

### Core Capabilities
- **OAuth2 Authentication** with device flow and PKCE
- **Dynamic Plugin System** for extensible functionality
- **Background Worker** for long-running tasks
- **Cross-Platform** support (Linux, macOS, Windows)
- **Enterprise Security** with secure token storage
- **Real-time Progress** tracking and streaming
- **Package Manager Integration** (Homebrew, APT, Chocolatey)

### Built-in Modules
- **YouTube Downloader** - Video/audio download with format selection
- **Format Listing** - Comprehensive format information
- **Video Information** - Metadata extraction and display

### Enterprise Features
- **Device Management** - Secure device registration and revocation
- **Audit Logging** - Complete activity tracking
- **Token Management** - Automatic refresh and rotation
- **Custom Domains** - Enterprise-ready domain configuration
- **SSO Integration** - Ready for enterprise identity providers

## 📦 Installation

### Quick Install

#### Linux/macOS
```bash
# Download and install
curl -L https://cli.conversoempire.world/install.sh | bash

# Or using Homebrew
brew install converso-empire/tap/converso-cli
```

#### Windows
```powershell
# Using Chocolatey
choco install converso-cli

# Or manual download
Invoke-WebRequest -Uri "https://cli.conversoempire.world/converso-windows-amd64.zip" -OutFile "converso.zip"
Expand-Archive -Path "converso.zip" -DestinationPath "$env:ProgramFiles\Converso"
```

### From Source
```bash
# Clone repository
git clone https://github.com/converso-empire/cli.git
cd cli

# Build
make build

# Install
sudo make install
```

## 🔧 Quick Start

### 1. Setup
```bash
# Initial setup
converso setup

# Check system requirements
converso setup --verbose
```

### 2. Authentication
```bash
# Login with OAuth2 device flow
converso login

# Check authentication status
converso status

# Logout
converso logout
```

### 3. YouTube Module
```bash
# List available formats
converso youtube list-formats https://youtube.com/watch?v=example

# Download video
converso youtube download https://youtube.com/watch?v=example

# Download audio only
converso youtube download https://youtube.com/watch?v=example --mode audio

# Get video information
converso youtube info https://youtube.com/watch?v=example
```

## 🏗️ Architecture

### Hybrid Design
```
┌─────────────────┐    JSON/gRPC    ┌─────────────────┐
│   Go CLI Shell  │◄───────────────►│ Python Engine   │
│                 │                 │                 │
│ • Commands      │                 │ • YouTube       │
│ • Auth          │                 │ • Format Logic  │
│ • Device Mgmt   │                 │ • Progress      │
│ • Background    │                 │ • Backend Sync  │
│ • Plugin Router │                 │ • FFmpeg        │
└─────────────────┘                 └─────────────────┘
```

### Key Components

#### Go CLI Shell
- **Command Framework**: Cobra-based CLI with subcommands
- **Authentication**: OAuth2 with device flow and PKCE
- **Plugin System**: Dynamic Python module loading
- **Background Worker**: Job queue and task processing
- **Configuration**: Viper-based configuration management

#### Python Engine
- **Module Bridge**: JSON IPC communication layer
- **YouTube Module**: Wraps existing yt-dlp functionality
- **Progress Streaming**: Real-time progress event emission
- **Backend Sync**: Activity logging and status reporting

## 📚 Usage

### Authentication
```bash
# Login with custom device name
converso login --device-name "My Laptop"

# Force re-authentication
converso login --force

# Check status
converso status
```

### YouTube Commands
```bash
# Download with specific format
converso youtube download <url> --format-id 140 --container m4a

# Download to custom directory
converso youtube download <url> --output-dir ./downloads

# List formats with details
converso youtube list-formats <url>

# Get video metadata
converso youtube info <url>
```

### Plugin Management
```bash
# List installed plugins
converso plugins list

# Install new plugin
converso plugins install <plugin-name> <source>

# Update plugin
converso plugins update <plugin-name>

# Remove plugin
converso plugins remove <plugin-name>
```

### Background Jobs
```bash
# Start background worker
converso worker start

# Check worker status
converso worker status

# Stop worker
converso worker stop
```

## 🔐 Security

### Authentication Flow
1. **Device Registration**: Unique device ID generation
2. **OAuth2 Device Flow**: Secure authentication with PKCE
3. **Token Storage**: Encrypted storage using OS keychain
4. **Automatic Refresh**: Seamless token rotation
5. **Device Revocation**: Secure logout and cleanup

### Security Features
- **Token Encryption**: AES-256 encryption for stored tokens
- **Secure IPC**: JSON-based communication with validation
- **Input Validation**: Comprehensive input sanitization
- **Audit Logging**: Complete activity tracking
- **Rate Limiting**: Protection against abuse

## 🧩 Plugin System

### Creating Plugins

#### Plugin Structure
```
plugins/
└── my-module/
    ├── manifest.json
    ├── __main__.py
    └── requirements.txt
```

#### Plugin Manifest
```json
{
  "name": "my-module",
  "version": "1.0.0",
  "description": "My custom module",
  "commands": ["command1", "command2"],
  "dependencies": ["requests", "click"],
  "author": "Your Name",
  "license": "MIT"
}
```

#### Plugin Implementation
```python
#!/usr/bin/env python3
from bridge import ModuleBase

class MyModule(ModuleBase):
    def __init__(self):
        super().__init__()
        self.register_command("command1", self.command1)
        self.register_command("command2", self.command2)
    
    def command1(self, args):
        return {"message": "Command 1 executed", "args": args}
    
    def command2(self, args):
        return {"message": "Command 2 executed", "args": args}

def main():
    module = MyModule()
    module.run()

if __name__ == "__main__":
    main()
```

### Plugin Commands
```bash
# Install plugin from local path
converso plugins install my-module ./path/to/plugin

# Install plugin from URL
converso plugins install my-module https://example.com/plugin.zip

# List installed plugins
converso plugins list

# Show plugin info
converso plugins info my-module
```

## 🚀 Development

### Prerequisites
- Go 1.21+
- Python 3.8+
- FFmpeg (for media processing)

### Building
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

### Testing
```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run benchmarks
make bench
```

### Code Quality
```bash
# Format code
make format

# Run linter
make lint

# Security scan
make security
```

### Development Workflow
```bash
# Setup development environment
make setup

# Start development with watch mode
make watch

# Run development server
make dev ARGS="--help"
```

## 📊 Monitoring

### Activity Logging
All CLI operations are logged with:
- Command execution
- Authentication events
- Plugin activity
- Error conditions
- Performance metrics

### Metrics Collection
- Command execution times
- Plugin performance
- Authentication success rates
- Error rates and types

### Health Monitoring
- Worker status
- Plugin health
- Authentication status
- System resource usage

## 🔧 Configuration

### Configuration File Location
- **Linux**: `~/.converso/config.yaml`
- **macOS**: `~/.converso/config.yaml`
- **Windows**: `%USERPROFILE%\.converso\config.yaml`

### Configuration Options
```yaml
# Debug mode
debug: false

# API Configuration
api_endpoint: "https://capi.conversoempire.world"
auth_url: "https://clerk.conversoempire.world/oauth/authorize"
token_url: "https://clerk.conversoempire.world/oauth/token"
client_id: "converso-cli"

# Application Settings
concurrency: 10
device_name: "default"

# Paths (auto-generated)
# data_dir: "~/.converso/data"
# plugins_dir: "~/.converso/plugins"
```

### Environment Variables
```bash
# Override configuration
export CONVERSO_DEBUG=true
export CONVERSO_API_ENDPOINT="https://custom.domain/api"
export CONVERSO_CLIENT_ID="custom-client"
```

## 🐛 Troubleshooting

### Common Issues

#### Authentication Problems
```bash
# Clear authentication and retry
converso logout --force
converso login
```

#### Plugin Issues
```bash
# Reinstall plugin
converso plugins remove youtube
converso plugins install youtube ./python-engine/modules/youtube
```

#### FFmpeg Not Found
```bash
# Install FFmpeg
# Linux (Ubuntu/Debian)
sudo apt-get install ffmpeg

# macOS
brew install ffmpeg

# Windows
# Download from https://ffmpeg.org/download.html
```

#### Python Path Issues
```bash
# Check Python installation
python3 --version

# Set Python path in config
converso config set python_path /usr/bin/python3
```

### Debug Mode
```bash
# Enable debug logging
converso --debug youtube list-formats <url>

# Check logs
tail -f ~/.converso/logs/converso.log
```

### Support
- **Documentation**: https://cli.conversoempire.world
- **Issues**: https://github.com/converso-empire/cli/issues
- **Discussions**: https://github.com/converso-empire/cli/discussions

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📞 Support

For enterprise support and custom deployments, please contact [support@conversoempire.world](mailto:support@conversoempire.world).

## 🔗 Links

- **Documentation**: https://cli.conversoempire.world
- **API Reference**: https://api.conversoempire.world/docs
- **GitHub**: https://github.com/converso-empire/cli
- **Issues**: https://github.com/converso-empire/cli/issues
- **Discussions**: https://github.com/converso-empire/cli/discussions

---

**Converso CLI** - Enterprise-grade command line interface for the modern developer.

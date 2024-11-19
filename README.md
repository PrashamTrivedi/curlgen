# 🌟 CurlGen CLI - AI-Powered API Testing Tool

## 📖 About
CurlGen is a sophisticated CLI tool that leverages AI models to automatically generate cURL commands for API testing. Built with Deno/TypeScript, it supports both OpenAI GPT and Anthropic Claude models to provide intelligent and context-aware command generation.

> 🎯 **Key Features**
> - AI-powered cURL command generation
> - Multiple AI model support (OpenAI & Anthropic)
> - Secure configuration management
> - Cross-platform compatibility

## 🚀 Installation and Usage

### Prerequisites
```bash
# Install Deno if you haven't already
curl -fsSL https://deno.land/x/install/install.sh | sh
```

### Setup
1. Clone the repository
2. Configure your AI provider API keys:
```bash
curlgen config set openai_key YOUR_KEY
curlgen config set anthropic_key YOUR_KEY
```

### Basic Usage
```bash
# Generate cURL commands
 curlgen --files serverless.yml --files src/handler.ts --files src/model.ts -m claude-3.5-sonnet -p "We are doing sanity for this API Gateway" -t "We need test curls to help us do the sanity test for this API Gateway" -e apiUrl -c false
```

## 💻 About the Code

### Core Components
| Component | Description |
|-----------|-------------|
| `aiClients.ts` | Manages AI model clients and handles model selection |
| `config.ts` | Handles secure configuration and API key management |
| `prompt.ts` | Processes AI interactions and command generation |
| `utils.ts` | Provides utility functions for file operations |

### Project Structure
```
📦 CurlGen
 ┣ 📂 .github/workflows      # CI/CD configuration
 ┣ 📄 aiClients.ts          # AI client management
 ┣ 📄 config.ts            # Configuration handling
 ┣ 📄 main.ts             # Application entry point
 ┣ 📄 prompt.ts           # Command generation logic
 ┗ 📄 utils.ts            # Utility functions
```

> 🔧 **Technical Highlights**
> - Lazy initialization of AI clients
> - Cross-platform configuration management
> - Modular architecture for easy extension
> - Comprehensive error handling
> - Automated testing and CI/CD integration
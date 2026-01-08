# 🤖 LLM to Anthropic Proxy

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=flat)
![Build](https://img.shields.io/badge/Build-Passing-brightgreen?style=flat)
![Release](https://img.shields.io/github/v/release/nerdneilsfield/llm-to-anthropic?style=flat&logo=github)
![Docker Hub](https://img.shields.io/docker/v/nerdneilsfield/llm-to-anthropic?style=flat&logo=docker)
![GHCR](https://img.shields.io/badge/ghcr.io-latest-blue?style=flat&logo=github)
![Issues](https://img.shields.io/github/issues/nerdneilsfield/llm-to-anthropic?style=flat)
![Forks](https://img.shields.io/github/forks/nerdneilsfield/llm-to-anthropic?style=flat)
![Stars](https://img.shields.io/github/stars/nerdneilsfield/llm-to-anthropic?style=flat)

**A flexible LLM API proxy that translates various LLM providers into a unified Anthropic-compatible format**

[Quick Start](#-quick-start) • [Configuration](#-configuration) • [API Docs](#-api-reference) • [Docker](#-docker--deployment) • [Examples](#-examples)

</div>

---

<div align="right">
  <a href="README_zh.md">🇨🇳 中文版</a>
</div>


---

## ✨ Features

- 🎯 **Multi-Provider Support** - Configure any number of LLM providers (OpenAI, Anthropic, Google Gemini, Ollama, etc.)
- 🔑 **Flexible API Keys** - Support direct keys, environment variables, or bypass mode
- 🔄 **Model Mappings** - Map simple names like `haiku` to any provider/model combination
- 🚀 **Client-Side Keys** - Forward client API keys to providers (bypass mode)
- ⚡ **High Performance** - Built with Fiber v2 and fasthttp for blazing speed
- 🛡️ **Configuration Validation** - Validate all settings at startup with clear error messages
- 📝 **Anthropic Compatible** - Drop-in replacement for Anthropic API

---

## 🚀 Quick Start

### Installation

Choose one of the following installation methods:

#### Method 1: Download Pre-built Binary (Recommended)

```bash
# Download the latest binary for your platform
# Linux AMD64
wget https://github.com/nerdneilsfield/llm-to-anthropic/releases/latest/download/llm-to-anthropic-linux-amd64 -O llm-to-anthropic

# macOS AMD64
wget https://github.com/nerdneilsfield/llm-to-anthropic/releases/latest/download/llm-to-anthropic-darwin-amd64 -O llm-to-anthropic

# Windows AMD64
wget https://github.com/nerdneilsfield/llm-to-anthropic/releases/latest/download/llm-to-anthropic-windows-amd64.exe -O llm-to-anthropic.exe

# Make executable (Linux/macOS)
chmod +x llm-to-anthropic

# Run
./llm-to-anthropic serve
```


#### Method 2: Using Go Install

```bash
# Install directly from GitHub
go install github.com/nerdneilsfield/llm-to-anthropic@latest

# The binary will be installed to $GOPATH/bin
# Add $GOPATH/bin to your PATH if not already added
export PATH=$PATH:$(go env GOPATH)/bin

# Run
llm-to-anthropic serve
```

#### Method 3: Using Docker

```bash
# Pull and run image
docker run -d \
  -p 8082:8082 \
  -v $(pwd)/config.toml:/app/config.toml \
  nerdneilsfield/llm-to-anthropic:latest

# Or use GitHub Container Registry
docker run -d \
  -p 8082:8082 \
  -v $(pwd)/config.toml:/app/config.toml \
  ghcr.io/nerdneilsfield/llm-to-anthropic:latest
```

#### Method 4: Build from Source


```bash
# Clone the repository
git clone https://github.com/nerdneilsfield/llm-to-anthropic.git
cd llm-to-anthropic

# Build from source
go build -o llm-to-anthropic .

# Run
./llm-to-anthropic serve
```

### Minimal Configuration

Create `config.toml`:

```toml
[server]
host = "0.0.0.0"
port = 8082

[[providers]]
name = "openai"
type = "openai"
api_base_url = "https://api.openai.com/v1"
api_key = "env:OPENAI_API_KEY"
models = ["gpt-4o", "gpt-4.1-mini"]
```

### Make Your First Request

```bash
curl -X POST http://localhost:8082/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-openai-api-key" \
  -d '{
    "model": "openai/gpt-4o",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

---

## 🐳 Docker & Deployment

For Docker usage and deployment guides, see:

- 📦 [Docker Usage Guide](DOCKER.md) - Run with Docker or Docker Compose
- 🚀 [Deployment Guide](DEPLOYMENT.md) - Release process, CI/CD, production deployment
- 🔐 [Security Best Practices](#-security-best-practices)

---

## 📖 Configuration

### Basic Structure

```toml
[server]
host = "0.0.0.0"
port = 8082
read_timeout = 120
write_timeout = 120

# Define multiple providers
[[providers]]
name = "openai"
type = "openai"
api_base_url = "https://api.openai.com/v1"
api_key = "env:OPENAI_API_KEY"
models = ["gpt-4o", "gpt-4.1-mini"]

[[providers]]
name = "ollama"
type = "openai"
api_base_url = "http://localhost:11434/v1"
api_key = "bypass"
models = ["llama3.2:3b", "llama3.2:7b"]

# Model mappings
[mappings]
"haiku" = "ollama/llama3.2:3b"
"sonnet" = "ollama/llama3.2:7b"
```

<details>
<summary><strong>🔧 Advanced Configuration Options</strong></summary>

### API Key Configuration

Three modes are supported:

#### 1. Direct Key
```toml
api_key = "sk-xxxxxxxxxxxxxxxx"
```
Write the API key directly in the config file.

#### 2. Environment Variable (Recommended)
```toml
api_key = "env:OPENAI_API_KEY"
```
Read from environment variable. The proxy will validate that the variable exists and is not empty at startup.

#### 3. Bypass/Forward Mode
```toml
api_key = "bypass"  # or "forward"
```
Forward the client's `X-API-Key` header to the provider. Useful when you want clients to manage their own keys.

### Provider Types

| Type | Description | Example |
|------|-------------|----------|
| `openai` | OpenAI-compatible API | OpenAI, Azure, Ollama, DeepSeek |
| `anthropic` | Anthropic API | Claude models |
| `gemini` | Google Gemini API | Gemini models |

### Model Selection

Use the `provider/model` format:

```bash
# Direct provider/model specification
curl -d '{"model": "openai/gpt-4o", ...}'

# Or use mappings
curl -d '{"model": "haiku", ...}'  # Maps to "ollama/llama3.2:3b"
```

### Vertex AI Configuration

For Google Vertex AI:

```toml
[[providers]]
name = "vertex"
type = "gemini"
api_base_url = "https://us-central1-aiplatform.googleapis.com/v1"
api_key = "bypass"
use_vertex_auth = true
vertex_project = "your-project-id"
vertex_location = "us-central1"
models = ["gemini-2.5-pro"]
```

### Configuration Validation

The proxy validates all settings at startup:

```bash
# Example validation errors
Failed to load configuration: invalid configuration: 
  provider openai: environment variable 'OPENAI_API_KEY' is not set or is empty

Failed to load configuration: invalid configuration: 
  provider openai: models list is required and must not be empty

Failed to load configuration: invalid configuration: 
  mapping: alias 'test' references non-existent provider 'nonexistent'
```

See [CONFIGURATION_VALIDATION.md](CONFIGURATION_VALIDATION.md) for complete validation rules.

</details>

---

## 📚 API Reference

### Health Check Endpoints

#### GET /health
Basic health check.

```bash
curl http://localhost:8082/health
```

**Response:**
```json
{
  "status": "ok"
}
```

#### GET /health/ready
Readiness check with provider status.

```bash
curl http://localhost:8082/health/ready
```

**Response:**
```json
{
  "status": "ready",
  "providers": {
    "openai": "configured",
    "ollama": "configured"
  },
  "total_providers": 2,
  "total_mappings": 2
}
```

### Message Endpoint

#### POST /v1/messages
Send messages using Anthropic API format.

```bash
curl -X POST http://localhost:8082/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "openai/gpt-4o",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `model` | string | Yes | Model identifier (e.g., `openai/gpt-4o`) |
| `max_tokens` | integer | Yes | Maximum tokens to generate |
| `messages` | array | Yes | Array of message objects |
| `stream` | boolean | No | Enable streaming (default: false) |

**Response:**
```json
{
  "id": "msg_123",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "Hello! How can I help you today?"
    }
  ],
  "model": "openai/gpt-4o",
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 10,
    "output_tokens": 20
  }
}
```

### Models Endpoint

#### GET /v1/models
List all available models.

```bash
curl http://localhost:8082/v1/models
```

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "id": "openai/gpt-4o",
      "object": "model",
      "created": 1234567890,
      "owned_by": "openai"
    },
    {
      "id": "ollama/llama3.2:3b",
      "object": "model",
      "created": 1234567890,
      "owned_by": "ollama"
    }
  ]
}
```

<details>
<summary><strong>🔧 Advanced API Usage</strong></summary>

### Streaming Responses

Enable streaming by setting `stream: true`:

```bash
curl -X POST http://localhost:8082/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "openai/gpt-4o",
    "max_tokens": 1024,
    "stream": true,
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

The response will be sent as Server-Sent Events (SSE).

### Error Responses

All errors follow the Anthropic API error format:

```json
{
  "type": "invalid_request_error",
  "error": {
    "type": "invalid_request_error",
    "message": "model field is required"
  }
}
```

### Rate Limiting

The proxy does not implement rate limiting. Rate limiting is handled by the upstream providers.

### Authentication

The proxy supports two authentication modes:

1. **Server-Side**: API keys are configured in `config.toml`
2. **Client-Side** (Bypass): Clients provide their own API keys via `X-API-Key` header

In bypass mode, the `X-API-Key` header is forwarded to the provider.

</details>

---

## 🎯 Examples

<details>
<summary><strong>📝 Example 1: Multiple Providers</strong></summary>

```toml
[[providers]]
name = "openai"
type = "openai"
api_base_url = "https://api.openai.com/v1"
api_key = "env:OPENAI_API_KEY"
models = ["gpt-4o", "gpt-4.1-mini"]

[[providers]]
name = "anthropic"
type = "anthropic"
api_base_url = "https://api.anthropic.com"
api_key = "env:ANTHROPIC_API_KEY"
models = ["claude-3-5-sonnet-20241022", "claude-haiku-4-20250514"]

[[providers]]
name = "ollama"
type = "openai"
api_base_url = "http://localhost:11434/v1"
api_key = "bypass"
models = ["llama3.2:3b", "llama3.2:7b"]
```

```bash
# Use OpenAI
curl -d '{"model": "openai/gpt-4o", ...}'

# Use Anthropic
curl -d '{"model": "anthropic/claude-3-5-sonnet-20241022", ...}'

# Use Ollama
curl -d '{"model": "ollama/llama3.2:7b", ...}'
```

</details>

<details>
<summary><strong>📝 Example 2: Model Mappings</strong></summary>

```toml
[mappings]
"haiku" = "ollama/llama3.2:1b"
"sonnet" = "ollama/llama3.2:3b"
"opus" = "ollama/llama3.2:7b"
"claude" = "anthropic/claude-3-5-sonnet-20241022"
"gpt" = "openai/gpt-4o"
```

```bash
# Simple names
curl -d '{"model": "haiku", ...}'   # Uses ollama/llama3.2:1b
curl -d '{"model": "sonnet", ...}'  # Uses ollama/llama3.2:3b
curl -d '{"model": "claude", ...}'  # Uses anthropic/claude-3-5-sonnet-20241022
```

</details>

<details>
<summary><strong>📝 Example 3: Custom OpenAI-Compatible API</strong></summary>

```toml
[[providers]]
name = "deepseek"
type = "openai"
api_base_url = "https://api.deepseek.com/v1"
api_key = "bypass"  # Let clients provide their own key
models = ["deepseek-chat", "deepseek-coder"]
```

```bash
curl -X POST http://localhost:8082/v1/messages \
  -H "x-api-key: your-deepseek-api-key" \
  -d '{"model": "deepseek/deepseek-chat", ...}'
```

</details>

<details>
<summary><strong>📝 Example 4: Local LLM with Ollama</strong></summary>

```toml
[[providers]]
name = "local"
type = "openai"
api_base_url = "http://localhost:11434/v1"
api_key = "bypass"
models = ["llama3.2:1b", "llama3.2:3b", "llama3.2:7b"]

[mappings]
"fast" = "local/llama3.2:1b"
"balanced" = "local/llama3.2:3b"
"powerful" = "local/llama3.2:7b"
```

```bash
# Use local LLM
curl -d '{"model": "local/llama3.2:7b", ...}'
curl -d '{"model": "powerful", ...}'  # Same as above
```

</details>

---

## 🛠️ Development

### Build from Source

```bash
# Clone repository
git clone https://github.com/nerdneilsfield/llm-to-anthropic.git
cd llm-to-anthropic

# Build
go build -o llm-to-anthropic .

# Run tests
go test ./...

# Run validation tests
./test_validation.sh
```

### Project Structure

```
llm-to-anthropic/
├── cmd/                # CLI commands
│   ├── proxy/         # Proxy command
│   └── root.go       # Root command
├── internal/          # Internal packages
│   ├── config/       # Configuration
│   ├── server/       # HTTP server
│   └── ...          # Other internal packages
├── pkg/              # Public packages
│   ├── provider/      # Provider clients
│   ├── api/          # API handlers
│   └── logger/       # Logging
├── config.toml       # Configuration file
├── README.md         # This file
└── main.go           # Entry point
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (use conventional commits: `feat:`, `fix:`, `docs:`, etc.)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🔒 Security Best Practices

1. **Never commit API keys** to your repository
2. **Use environment variables** for sensitive configuration
3. **Set proper file permissions** on config.toml (`chmod 600`)
4. **Use HTTPS** in production environments
5. **Keep images updated** to get security patches
6. **Review dependencies** regularly for vulnerabilities
7. **Use rate limiting** at the provider level
8. **Monitor logs** for suspicious activity
9. **Implement authentication** in reverse proxy if needed
10. **Regular backups** of configuration files

---

## 🤝 Support

- 📖 [Documentation](CONFIGURATION_VALIDATION.md)
- 🐛 [Issue Tracker](https://github.com/nerdneilsfield/llm-to-anthropic/issues)
- 💬 [Discussions](https://github.com/nerdneilsfield/llm-to-anthropic/discussions)
- 📦 [Releases](https://github.com/nerdneilsfield/llm-to-anthropic/releases)
- 🐳 [Docker Hub](https://hub.docker.com/r/nerdneilsfield/llm-to-anthropic)
- 📦 [GitHub Container Registry](https://github.com/nerdneilsfield/llm-to-anthropic/pkgs/container/llm-to-anthropic)

---

<div align="center">

**Made with ❤️ by the community**

[⬆ Back to Top](#-llm-to-anthropic-proxy)

</div>

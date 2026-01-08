# LLM API Proxy

A high-performance Go-based LLM API proxy that translates various LLM provider APIs (OpenAI, Google Gemini, Anthropic) into a unified Anthropic-compatible format.

## Features

- **Multiple Provider Support**: OpenAI, Google Gemini, and direct Anthropic API proxy
- **Anthropic API Compatible**: Full compatibility with Anthropic's v1 messages API
- **Streaming & Non-Streaming**: Full support for both streaming and non-streaming responses
- **Model Mapping**: Flexible model mapping configuration (haiku/sonnet → provider models)
- **High Performance**: Built on Fiber v2 and fasthttp for maximum throughput
- **Easy Configuration**: Simple environment variable configuration
- **Health Monitoring**: Built-in health check endpoints for orchestration systems
- **Flexible Authentication**: Support for both server-side and client-side API key management

## Quick Start

### Prerequisites

- Go 1.23.2 or later
- API keys for at least one provider (OpenAI, Google Gemini, or Anthropic) **OR** configure proxy to accept client-provided keys

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/llm-to-anthropic.git
cd llm-to-anthropic

# Build the project
make build

# Or build with just
just build
```

### Configuration

Create a `.env` file from the example:

```bash
cp .env.example .env
```

#### Authentication Options

**Option 1: Server-Side API Keys (Simplified)**

Configure API keys in `.env` - the proxy will handle all authentication:

```dotenv
# Configure API keys for each provider
OPENAI_API_KEY=sk-...
GEMINI_API_KEY=...
ANTHROPIC_API_KEY=sk-ant-...
```

Then client requests don't need API keys:

```bash
curl http://localhost:8082/v1/models
```

**Option 2: Client-Side API Keys (Multi-Tenant/Flexible)**

The proxy can forward API keys provided by clients. This is useful for:
- Multi-tenant applications where each user has their own keys
- Security (server never stores sensitive keys)
- Flexibility (different requests can use different keys)

Client provides API key via `X-Api-Key` header:

```bash
# Using OpenAI key
curl http://localhost:8082/v1/models \
  -H "X-Api-Key: sk-your-openai-key"

# Using Gemini key
curl http://localhost:8082/v1/models \
  -H "X-Api-Key: AIza-your-gemini-key"

# Using Anthropic key
curl http://localhost:8082/v1/models \
  -H "X-Api-Key: sk-ant-your-anthropic-key"
```

**How it works:**
- Client sends `X-Api-Key: <key>` header
- Proxy detects which provider the key is for (or uses model prefix)
- Proxy converts to provider format:
  - OpenAI: `Authorization: Bearer <key>`
  - Gemini: `?key=<key>` or Vertex AI ADC
  - Anthropic: `x-api-key: <key>`
- Proxy forwards request with converted header

**You can combine both approaches:**
- Server-side keys serve as fallback when client doesn't provide one
- Client-side keys override server defaults when provided

Edit `.env` with your configuration:

```dotenv
# Provider Configuration
PREFERRED_PROVIDER=openai  # or "google", "anthropic"

# Model Configuration (optional)
BIG_MODEL=gpt-4.1
SMALL_MODEL=gpt-4.1-mini

# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8082

# Logging
VERBOSE=false
```

### Running the Server

```bash
# Run the proxy server
./llm-to-anthropic proxy

# Or with verbose logging
./llm-to-anthropic proxy -v
```

## Usage with Claude Code

```bash
# Set ANTHROPIC_BASE_URL to point to your proxy
ANTHROPIC_BASE_URL=http://localhost:8082 claude
```

## Model Mapping

The proxy automatically maps Claude model names to provider-specific models:

| Claude Model | Default OpenAI | Default Google |
|--------------|----------------|----------------|
| haiku | gpt-4.1-mini | gemini-2.5-flash |
| sonnet | gpt-4.1 | gemini-2.5-pro |

You can also use explicit model names with provider prefixes:

- `openai/gpt-4o`
- `gemini/gemini-2.5-pro`
- `anthropic/claude-sonnet-4-20250514`

## API Endpoints

### Health Checks

- `GET /health` - Basic health check (always returns 200)
- `GET /health/ready` - Readiness check with provider status

### Anthropic API v1

- `POST /v1/messages` - Send messages to LLM
- `GET /v1/models` - List available models

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|-----------|-------------|
| `OPENAI_API_KEY` | No | OpenAI API key (default/fallback, can be provided by client) |
| `GEMINI_API_KEY` | No | Google Gemini API key (default/fallback, can be provided by client) |
| `ANTHROPIC_API_KEY` | No | Anthropic API key (default/fallback, can be provided by client) |
| `USE_VERTEX_AUTH` | No | Use Vertex AI ADC instead of API key (default: false) |
| `VERTEX_PROJECT` | No | Google Cloud Project ID (required with Vertex auth) |
| `VERTEX_LOCATION` | No | Google Cloud region (required with Vertex auth) |
| `PREFERRED_PROVIDER` | No | Preferred provider (default: openai) |
| `BIG_MODEL` | No | Model for "sonnet" requests |
| `SMALL_MODEL` | No | Model for "haiku" requests |
| `SERVER_HOST` | No | Server host (default: 0.0.0.0) |
| `SERVER_PORT` | No | Server port (default: 8082) |
| `VERBOSE` | No | Enable verbose logging (default: false) |

### Client Authentication

Clients can authenticate in two ways:

1. **Server-Side Keys**: Configure API keys in `.env`, clients don't need to provide keys
2. **Client-Side Keys**: Provide API key via `X-Api-Key` header, supports per-request keys

Example with client-provided key:

```bash
curl http://localhost:8082/v1/messages \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: sk-ant-your-key" \
  -d '{
    "model": "sonnet",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Supported Providers

#### OpenAI
- Full chat completion API support
- All GPT-4 and GPT-3.5 models
- Streaming support

#### Google Gemini
- Both API key and Vertex AI authentication
- Gemini 2.5 Pro and Flash models
- Streaming support

#### Anthropic
- Direct proxy mode (no translation)
- All Claude models
- Full API compatibility

## Development

### Project Structure

```
├── cmd/
│   └── proxy/          # CLI command
├── internal/
│   ├── config/         # Configuration management
│   └── server/        # HTTP server
├── pkg/
│   ├── api/proxy/     # API translation layer
│   │   ├── anthropic/ # Anthropic types
│   │   ├── openai/    # OpenAI translation
│   │   └── gemini/    # Gemini translation
│   ├── provider/       # Backend clients
│   │   ├── openai/
│   │   ├── gemini/
│   │   └── anthropic/
│   └── logger/        # Logger utility
└── openspec/          # OpenSpec specifications
```

### Building

```bash
# Build all
make build

# Build for specific platform
GOOS=linux GOARCH=amd64 make build

# Build for multiple platforms
just build-all
```

### Testing

```bash
# Run tests
make test

# Run with coverage
make cover

# Run specific package
go test ./pkg/api/proxy/openai/
```

### Linting

```bash
# Run linter
make lint

# Format code
make fmt
```

## Deployment

### Docker

```dockerfile
FROM golang:1.23.2-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o llm-to-anthropic .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/llm-to-anthropic .
COPY .env.example .
EXPOSE 8082
CMD ["./llm-to-anthropic", "proxy"]
```

### Docker Compose

```yaml
version: '3.8'
services:
  proxy:
    build: .
    ports:
      - "8082:8082"
    env_file:
      - .env
    restart: unless-stopped
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: llm-proxy
spec:
  replicas: 3
  selector:
    matchLabels:
      app: llm-proxy
  template:
    metadata:
      labels:
        app: llm-proxy
    spec:
      containers:
      - name: proxy
        image: llm-proxy:latest
        ports:
        - containerPort: 8082
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: openai
        - name: GEMINI_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: gemini
        livenessProbe:
          httpGet:
            path: /health
            port: 8082
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8082
```

## Performance

Built with performance in mind:

- **Fiber v2**: High-performance HTTP framework based on fasthttp
- **fasthttp**: Fast HTTP client for provider communication
- **Zero-copy**: Minimal memory allocations in hot paths
- **Streaming**: Efficient streaming response handling

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Inspired by [claude-code-proxy](https://github.com/1rgs/claude-code-proxy)
- Built with [Fiber v2](https://docs.gofiber.io/)
- Uses [fasthttp](https://github.com/valyala/fasthttp)

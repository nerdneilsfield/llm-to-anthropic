# LLM to Anthropic Proxy

A flexible LLM API proxy that translates various LLM provider APIs (OpenAI, Google Gemini, Anthropic) into a unified Anthropic-compatible format.

## Features

- **Multiple Providers**: Configure any number of LLM providers
- **Flexible API Keys**: Support direct keys, environment variables, or bypass mode
- **Model Mappings**: Map model names to provider/model combinations
- **Claude Compatibility**: Use Claude model names (haiku, sonnet, opus) with any provider
- **Client-side Keys**: Forward client API keys to providers (bypass mode)

## Installation

```bash
# Build from source
go build -o llm-to-anthropic .

# Run
./llm-to-anthropic serve
```

## Configuration

Create a `config.toml` file:

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
api_key = "env:OPENAI_API_KEY"  # Read from environment variable
models = [
    "gpt-4.1-mini",
    "gpt-4o",
]

[[providers]]
name = "ollama"
type = "openai"
api_base_url = "http://localhost:11434/v1"
api_key = "bypass"  # Forward client key
models = [
    "llama3.2:1b",
    "llama3.2:3b",
]

[[providers]]
name = "anthropic"
type = "anthropic"
api_base_url = "https://api.anthropic.com"
api_key = "sk-ant-xxx"  # Direct key
models = [
    "claude-haiku-4-20250514",
    "claude-3-5-sonnet-20241022",
]

# Model mappings
[mappings]
"haiku" = "ollama/llama3.2:1b"
"sonnet" = "ollama/llama3.2:3b"
"opus" = "ollama/llama3.2:7b"
```

### API Key Configuration

Three modes supported:

1. **Direct Key**: Write the API key directly in config
   ```toml
   api_key = "sk-xxx"
   ```

2. **Environment Variable**: Read from environment variable
   ```toml
   api_key = "env:OPENAI_API_KEY"
   ```

3. **Bypass/Forward**: Forward client's X-API-Key header
   ```toml
   api_key = "bypass"  # or "forward"
   ```

### Model Selection

Use `provider/model` format:

```bash
# Use OpenAI
curl -X POST /v1/messages \
  -d '{"model": "openai/gpt-4o", ...}'

# Use Ollama
curl -X POST /v1/messages \
  -d '{"model": "ollama/llama3.2:7b", ...}'

# Use Anthropic
curl -X POST /v1/messages \
  -d '{"model": "anthropic/claude-sonnet-4-20250514", ...}'
```

Or use mappings:

```bash
# Uses mapping "haiku" = "ollama/llama3.2:1b"
curl -X POST /v1/messages \
  -d '{"model": "haiku", ...}'
```

## Usage

### Start Server

```bash
# With default config
llm-to-anthropic serve

# With custom config path
CONFIG_PATH=/path/to/config.toml llm-to-anthropic serve

# Verbose logging
llm-to-anthropic serve -v
```

### Health Check

```bash
curl http://localhost:8082/health
curl http://localhost:8082/health/ready
```

### List Models

```bash
curl http://localhost:8082/v1/models
```

## API Endpoints

### POST /v1/messages

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

## License

MIT

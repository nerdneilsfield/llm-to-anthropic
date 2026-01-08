# Docker Usage

This guide explains how to build and run the LLM to Anthropic Proxy using Docker.

## 🐳 Quick Start

### Using Pre-built Images

#### Docker Hub
```bash
# Pull the image
docker pull nerdneilsfield/llm-to-anthropic:latest

# Run with default configuration
docker run -d -p 8082:8082 nerdneilsfield/llm-to-anthropic

# Run with custom configuration
docker run -d \
  -p 8082:8082 \
  -v $(pwd)/config.toml:/app/config.toml \
  nerdneilsfield/llm-to-anthropic

# Run with environment variables
docker run -d \
  -p 8082:8082 \
  -e OPENAI_API_KEY=sk-xxx \
  -e ANTHROPIC_API_KEY=sk-ant-xxx \
  nerdneilsfield/llm-to-anthropic
```

#### GitHub Container Registry
```bash
# Pull the image
docker pull ghcr.io/nerdneilsfield/llm-to-anthropic:latest

# Run
docker run -d -p 8082:8082 ghcr.io/nerdneilsfield/llm-to-anthropic
```

### Using Specific Version

```bash
# Tagged version
docker pull nerdneilsfield/llm-to-anthropic:v1.0.0

# Run with version
docker run -d -p 8082:8082 nerdneilsfield/llm-to-anthropic:v1.0.0
```

## 🔧 Building from Source

### Build Using Dockerfile
```bash
# Build the image
docker build -t llm-to-anthropic .

# Run the container
docker run -d -p 8082:8082 llm-to-anthropic
```

### Build Using GoReleaser
```bash
# Build multi-platform images
goreleaser release --clean --skip=publish

# Load image (for local testing)
docker load < llm-to-anthropic_linux_amd64.docker
```

## 📝 Configuration

### Using config.toml

Mount your configuration file:

```bash
docker run -d \
  -p 8082:8082 \
  -v $(pwd)/config.toml:/app/config.toml \
  nerdneilsfield/llm-to-anthropic
```

### Using Environment Variables

```bash
docker run -d \
  -p 8082:8082 \
  -e OPENAI_API_KEY=sk-xxx \
  -e ANTHROPIC_API_KEY=sk-ant-xxx \
  -e GEMINI_API_KEY=AIzaSyD-xxx \
  nerdneilsfield/llm-to-anthropic
```

### Combining Both

```bash
docker run -d \
  -p 8082:8082 \
  -v $(pwd)/config.toml:/app/config.toml \
  -e OPENAI_API_KEY=sk-xxx \
  nerdneilsfield/llm-to-anthropic
```

## 🌐 Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  llm-to-anthropic:
    image: nerdneilsfield/llm-to-anthropic:latest
    # Or use: ghcr.io/nerdneilsfield/llm-to-anthropic:latest
    container_name: llm-to-anthropic
    ports:
      - "8082:8082"
    volumes:
      # Mount configuration file
      - ./config.toml:/app/config.toml
    environment:
      # Optional: Override with environment variables
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8082/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

### Using Docker Compose

```bash
# Start service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop service
docker-compose down

# Restart service
docker-compose restart
```

## 🚀 Advanced Usage

### Custom Configuration File Location

```bash
docker run -d \
  -p 8082:8082 \
  -v /path/to/custom/config.toml:/app/config.toml \
  -e CONFIG_PATH=/app/config.toml \
  nerdneilsfield/llm-to-anthropic
```

### Using Ollama Sidecar

```yaml
version: '3.8'

services:
  ollama:
    image: ollama/ollama:latest
    container_name: ollama
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
    restart: unless-stopped

  llm-to-anthropic:
    image: nerdneilsfield/llm-to-anthropic:latest
    container_name: llm-to-anthropic
    ports:
      - "8082:8082"
    depends_on:
      - ollama
    volumes:
      - ./config.toml:/app/config.toml
    environment:
      - OPENAI_API_KEY=not-needed
    restart: unless-stopped

volumes:
  ollama_data:
```

### Production Deployment with Nginx

```yaml
version: '3.8'

services:
  llm-to-anthropic:
    image: nerdneilsfield/llm-to-anthropic:latest
    container_name: llm-to-anthropic
    ports:
      - "127.0.0.1:8082:8082"
    volumes:
      - ./config.toml:/app/config.toml
      - ./logs:/app/logs
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    container_name: nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - llm-to-anthropic
    restart: unless-stopped
```

## 🔍 Health Check

```bash
# Check container health
docker ps
docker inspect <container_id> | grep -A 10 Health

# Manual health check
curl http://localhost:8082/health

# Ready check
curl http://localhost:8082/health/ready
```

## 📊 Resource Limits

```bash
docker run -d \
  -p 8082:8082 \
  --memory="512m" \
  --memory-swap="1g" \
  --cpus="1.0" \
  -v $(pwd)/config.toml:/app/config.toml \
  nerdneilsfield/llm-to-anthropic
```

## 🐛 Troubleshooting

### Container Won't Start

```bash
# Check logs
docker logs llm-to-anthropic

# Common issues:
# 1. Configuration file syntax error
# 2. Port already in use
# 3. Environment variables not set
```

### Can't Connect to Service

```bash
# Check if container is running
docker ps

# Check port mapping
docker port llm-to-anthropic

# Test connectivity
curl http://localhost:8082/health
```

### Permission Issues

```bash
# Ensure config file is readable
chmod 644 config.toml

# Check file ownership
ls -la config.toml
```

## 📚 Supported Platforms

- `linux/amd64` - Most common
- `linux/arm64` - Apple Silicon, ARM servers

## 🔄 Image Tags

- `latest` - Latest stable release
- `vX.Y.Z` - Specific version (e.g., `v1.0.0`)
- `nightly` - Latest development build (if available)

## 📦 Multi-Architecture Support

The Docker images support multiple architectures. Docker will automatically pull the correct image for your platform.

```bash
# View image architectures
docker buildx imagetools inspect nerdneilsfield/llm-to-anthropic:latest
```

## 🔐 Security Best Practices

1. **Never commit API keys** to your repository
2. **Use environment variables** for sensitive data
3. **Mount read-only** configuration files when possible
4. **Run as non-root** (the image already does this)
5. **Use resource limits** in production
6. **Keep images updated** by pulling latest tags

```bash
# Example: Secure deployment
docker run -d \
  -p 8082:8082 \
  --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=100m \
  --security-opt=no-new-privileges \
  --cap-drop ALL \
  --cap-add NET_BIND_SERVICE \
  -e OPENAI_API_KEY=${OPENAI_API_KEY} \
  nerdneilsfield/llm-to-anthropic
```

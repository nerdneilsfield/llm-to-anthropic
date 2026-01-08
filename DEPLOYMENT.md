# Deployment Guide

This guide explains how to deploy and release the LLM to Anthropic Proxy.

## 🚀 Release Process

### Automatic Releases with GoReleaser

Releases are automatically triggered when you push a new tag:

```bash
# Create and push a new tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

This will:
1. Build binaries for multiple platforms (linux, darwin, windows, freebsd)
2. Create Docker images (linux/amd64, linux/arm64)
3. Push to Docker Hub (`nerdneilsfield/llm-to-anthropic`)
4. Push to GitHub Container Registry (`ghcr.io/nerdneilsfield/llm-to-anthropic`)
5. Create GitHub release with checksums
6. Upload artifacts to GitHub release

### Manual Testing Before Release

```bash
# Test build locally
go build -o llm-to-anthropic .
./llm-to-anthropic serve

# Test Docker build
docker build -t llm-to-anthropic .
docker run -p 8082:8082 llm-to-anthropic
```

## 🔐 Required GitHub Secrets

Configure these secrets in your repository settings (`https://github.com/nerdneilsfield/llm-to-anthropic/settings/secrets`):

### Docker Hub

| Secret Name | Description | How to Get |
|-------------|-------------|--------------|
| `DOCKERHUB_USERNAME` | Your Docker Hub username | Your Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub access token | Create in Docker Hub account settings |

#### Creating Docker Hub Token

1. Go to [Docker Hub Account Settings](https://hub.docker.com/settings/security)
2. Click "New Access Token"
3. Give it a description (e.g., "GitHub Actions")
4. Select permissions: `Read, Write, Delete`
5. Click "Generate"
6. Copy the token and add to GitHub Secrets as `DOCKERHUB_TOKEN`
7. Add your Docker Hub username as `DOCKERHUB_USERNAME`

### GitHub Container Registry

| Secret Name | Description | How to Get |
|-------------|-------------|--------------|
| `GITHUB_TOKEN` | Automatically provided by GitHub Actions | No action needed |

The `GITHUB_TOKEN` is automatically provided by GitHub Actions and has `packages: write` permission (configured in `.github/workflows/goreleaser.yml`).

## 📦 Artifacts

### Binaries

For each release, the following binaries are built:

| OS | Arch | Binary |
|-----|-------|---------|
| Linux | 386, amd64, arm64 | `llm-to-anthropic-linux-386`, etc. |
| Darwin | 386, amd64, arm64 | `llm-to-anthropic-darwin-amd64`, etc. |
| Windows | 386, amd64, arm64 | `llm-to-anthropic-windows-amd64.exe`, etc. |
| FreeBSD | 386, amd64, arm64 | `llm-to-anthropic-freebsd-amd64`, etc. |

### Docker Images

Each release publishes Docker images to two registries:

#### Docker Hub
- `nerdneilsfield/llm-to-anthropic:latest`
- `nerdneilsfield/llm-to-anthropic:vX.Y.Z`

#### GitHub Container Registry
- `ghcr.io/nerdneilsfield/llm-to-anthropic:latest`
- `ghcr.io/nerdneilsfield/llm-to-anthropic:vX.Y.Z`

#### Platforms
Both registries support:
- `linux/amd64`
- `linux/arm64`

## 🌍 Deployment Options

### Docker

```bash
# Pull and run
docker pull nerdneilsfield/llm-to-anthropic:latest
docker run -d -p 8082:8082 nerdneilsfield/llm-to-anthropic
```

See [DOCKER.md](DOCKER.md) for detailed Docker usage.

### Binary

```bash
# Download binary for your platform
wget https://github.com/nerdneilsfield/llm-to-anthropic/releases/download/v1.0.0/llm-to-anthropic-linux-amd64

# Make executable
chmod +x llm-to-anthropic-linux-amd64

# Run
./llm-to-anthropic-linux-amd64 serve
```

### Docker Compose

```yaml
version: '3.8'

services:
  llm-to-anthropic:
    image: nerdneilsfield/llm-to-anthropic:latest
    ports:
      - "8082:8082"
    volumes:
      - ./config.toml:/app/config.toml
    restart: unless-stopped
```

```bash
# Start
docker-compose up -d
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: llm-to-anthropic
spec:
  replicas: 3
  selector:
    matchLabels:
      app: llm-to-anthropic
  template:
    metadata:
      labels:
        app: llm-to-anthropic
    spec:
      containers:
      - name: llm-to-anthropic
        image: ghcr.io/nerdneilsfield/llm-to-anthropic:latest
        ports:
        - containerPort: 8082
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: llm-secrets
              key: openai-api-key
        - name: ANTHROPIC_API_KEY
          valueFrom:
            secretKeyRef:
              name: llm-secrets
              key: anthropic-api-key
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8082
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: llm-to-anthropic
spec:
  selector:
    app: llm-to-anthropic
  ports:
    - protocol: TCP
      port: 8082
      targetPort: 8082
  type: LoadBalancer
```

## 🔧 Configuration for Production

### Environment Variables

Use environment variables for production:

```bash
# Create .env file
cat > .env << EOF
OPENAI_API_KEY=sk-xxxxxxxx
ANTHROPIC_API_KEY=sk-ant-xxxxxxxx
GEMINI_API_KEY=AIzaSyD-xxxxxxxx
EOF

# Load and run
docker run -d \
  -p 8082:8082 \
  --env-file .env \
  -v $(pwd)/config.toml:/app/config.toml \
  nerdneilsfield/llm-to-anthropic
```

### Nginx Reverse Proxy

```nginx
upstream llm-to-anthropic {
    server localhost:8082;
}

server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://llm-to-anthropic;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### SSL/TLS with Let's Encrypt

```bash
# Use certbot
certbot certonly --standalone -d your-domain.com

# Configure Nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    
    location / {
        proxy_pass http://llm-to-anthropic;
        # ... headers
    }
}
```

## 📊 Monitoring

### Health Checks

```bash
# Basic health
curl http://localhost:8082/health

# Ready check with provider status
curl http://localhost:8082/health/ready
```

### Logging

```bash
# View container logs
docker logs -f llm-to-anthropic

# View logs since specific time
docker logs --since 1h llm-to-anthropic

# View logs with timestamps
docker logs -t llm-to-anthropic
```

### Metrics (Future)

Add Prometheus metrics endpoint at `/metrics` (planned feature).

## 🔒 Security

### Best Practices

1. **Use Secrets** for API keys
2. **Enable HTTPS** in production
3. **Configure Firewall** to only expose necessary ports
4. **Regular Updates** to get security patches
5. **Resource Limits** to prevent abuse
6. **Rate Limiting** (configure at provider level)

### Docker Security

```bash
# Run with security options
docker run -d \
  -p 8082:8082 \
  --read-only \
  --security-opt=no-new-privileges \
  --cap-drop ALL \
  --cap-add NET_BIND_SERVICE \
  -v $(pwd)/config.toml:/app/config.toml:ro \
  nerdneilsfield/llm-to-anthropic
```

## 🐛 Troubleshooting

### Release Fails

```bash
# Check GitHub Actions logs
# https://github.com/nerdneilsfield/llm-to-anthropic/actions

# Common issues:
# 1. DOCKERHUB_TOKEN not set or invalid
# 2. GITHUB_TOKEN missing packages:write permission
# 3. Tag doesn't follow semver (vX.Y.Z)
```

### Docker Push Fails

```bash
# Check if logged in
docker login

# Manually test push
docker push nerdneilsfield/llm-to-anthropic:test
```

### Container Won't Start

```bash
# Check logs
docker logs llm-to-anthropic

# Check configuration
docker run --rm \
  -v $(pwd)/config.toml:/app/config.toml \
  nerdneilsfield/llm-to-anthropic serve -v
```

## 📝 Release Checklist

Before releasing:

- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version number updated
- [ ] Docker Hub token configured
- [ ] GitHub Actions permissions verified
- [ ] Manual testing completed
- [ ] Security scan passed (if applicable)

After releasing:

- [ ] Verify binaries download
- [ ] Test Docker images
- [ ] Check GitHub release page
- [ ] Update website/docs (if applicable)
- [ ] Announce on social media/channels

## 🔄 Rollback Plan

If a release has issues:

```bash
# Pull previous version
docker pull nerdneilsfield/llm-to-anthropic:v0.9.0

# Update deployment
docker stop llm-to-anthropic
docker rm llm-to-anthropic
docker run -d -p 8082:8082 nerdneilsfield/llm-to-anthropic:v0.9.0
```

For Kubernetes:

```bash
kubectl rollout undo deployment/llm-to-anthropic
```

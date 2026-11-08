# 🤖 LLM 到 Anthropic 代理

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

**一个灵活的 LLM API 代理，将各种 LLM 提供商转换为统一的 Anthropic 兼容格式**

[快速开始](#-快速开始) • [配置](#-配置) • [API 文档](#-api-参考) • [Docker](#-docker-和部署) • [示例](#-示例)

</div>

---

<div align="right">
  <a href="README.md">🇺🇸 English Version</a>
</div>


---

## ✨ 特性

- 🎯 **多提供商支持** - 配置任意数量的 LLM 提供商（OpenAI、Anthropic、Google Gemini、Ollama 等）
- 🔑 **灵活的 API Key** - 支持直接 key、环境变量或绕过模式
- 🔄 **模型映射** - 将简单名称如 `haiku` 映射到任何提供商/模型组合
- 🚀 **客户端 Key** - 将客户端 API Key 转发到提供商（绕过模式）
- ⚡ **高性能** - 使用 Fiber v2 和 fasthttp 构建，速度飞快
- 🛡️ **配置验证** - 启动时验证所有设置，提供清晰的错误消息
- 📝 **Anthropic 兼容** - Anthropic API 的直接替代品

---

## 🚀 快速开始

### 通过 Homebrew / Scoop 安装

```bash
# macOS / Linux
brew install --cask nerdneilsfield/apps/llm-to-anthropic
```

```powershell
# Windows
scoop bucket add nerdneilsfield https://github.com/nerdneilsfield/homebrew-apps
scoop install nerdneilsfield/llm-to-anthropic
```

### 安装

选择以下任一安装方式：

#### 方式 1：下载预构建二进制（推荐）

```bash
# 下载适用于您平台的最新二进制文件
# Linux AMD64
wget https://github.com/nerdneilsfield/llm-to-anthropic/releases/latest/download/llm-to-anthropic-linux-amd64 -O llm-to-anthropic

# macOS AMD64
wget https://github.com/nerdneilsfield/llm-to-anthropic/releases/latest/download/llm-to-anthropic-darwin-amd64 -O llm-to-anthropic

# Windows AMD64
wget https://github.com/nerdneilsfield/llm-to-anthropic/releases/latest/download/llm-to-anthropic-windows-amd64.exe -O llm-to-anthropic.exe

# 添加执行权限（Linux/macOS）
chmod +x llm-to-anthropic

# 运行
./llm-to-anthropic serve
```


#### 方式 2：使用 Go Install

```bash
# 直接从 GitHub 安装
go install github.com/nerdneilsfield/llm-to-anthropic@latest

# 二进制文件将安装到 $GOPATH/bin
# 如果尚未添加，请将 $GOPATH/bin 添加到您的 PATH
export PATH=$PATH:$(go env GOPATH)/bin

# 运行
llm-to-anthropic serve
```

#### 方式 3：使用 Docker

```bash
# 拉取并运行镜像
docker run -d \
  -p 8082:8082 \
  -v $(pwd)/config.toml:/app/config.toml \
  nerdneilsfield/llm-to-anthropic:latest

# 或使用 GitHub Container Registry
docker run -d \
  -p 8082:8082 \
  -v $(pwd)/config.toml:/app/config.toml \
  ghcr.io/nerdneilsfield/llm-to-anthropic:latest
```

#### 方式 4：从源码构建


```bash
# 克隆仓库
git clone https://github.com/nerdneilsfield/llm-to-anthropic.git
cd llm-to-anthropic

# 从源码构建
go build -o llm-to-anthropic .

# 运行
./llm-to-anthropic serve
```

### 最小化配置

创建 `config.toml`:

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

### 发起第一个请求

```bash
curl -X POST http://localhost:8082/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: 你的-openai-api-key" \
  -d '{
    "model": "openai/gpt-4o",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "你好！"}
    ]
  }'
```

---

## 🐳 Docker 和部署

Docker 使用和部署指南：

- 📦 [Docker 使用指南](DOCKER.md) - 使用 Docker 或 Docker Compose 运行
- 🚀 [部署指南](DEPLOYMENT.md) - 发布流程、CI/CD、生产环境部署
- 🔐 [安全最佳实践](#-安全最佳实践)

---

## 📖 配置

### 基本结构

```toml
[server]
host = "0.0.0.0"
port = 8082
read_timeout = 120
write_timeout = 120

# 定义多个提供商
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

# 模型映射
[mappings]
"haiku" = "ollama/llama3.2:3b"
"sonnet" = "ollama/llama3.2:7b"
```

<details>
<summary><strong>🔧 高级配置选项</strong></summary>

### API Key 配置

支持三种模式：

#### 1. 直接 Key
```toml
api_key = "sk-xxxxxxxxxxxxxxxx"
```
直接在配置文件中写入 API key。

#### 2. 环境变量（推荐）
```toml
api_key = "env:OPENAI_API_KEY"
```
从环境变量读取。代理会在启动时验证该变量存在且不为空。

#### 3. 绕过/转发模式
```toml
api_key = "bypass"  # 或 "forward"
```
将客户端的 `X-API-Key` 请求头转发给提供商。适用于希望客户端管理自己的 key 的场景。

### 提供商类型

| 类型 | 描述 | 示例 |
|------|-------------|----------|
| `openai` | OpenAI 兼容 API | OpenAI、Azure、Ollama、DeepSeek |
| `anthropic` | Anthropic API | Claude 模型 |
| `gemini` | Google Gemini API | Gemini 模型 |

### 模型选择

使用 `provider/model` 格式：

```bash
# 直接指定 provider/model
curl -d '{"model": "openai/gpt-4o", ...}'

# 或使用映射
curl -d '{"model": "haiku", ...}'  # 映射到 "ollama/llama3.2:3b"
```

### Vertex AI 配置

对于 Google Vertex AI：

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

### 配置验证

代理在启动时验证所有设置：

```bash
# 示例验证错误
Failed to load configuration: invalid configuration: 
  provider openai: environment variable 'OPENAI_API_KEY' is not set or is empty

Failed to load configuration: invalid configuration: 
  provider openai: models list is required and must not be empty

Failed to load configuration: invalid configuration: 
  mapping: alias 'test' references non-existent provider 'nonexistent'
```

查看 [CONFIGURATION_VALIDATION.md](CONFIGURATION_VALIDATION.md) 了解完整的验证规则。

</details>

---

## 📚 API 参考

### 健康检查端点

#### GET /health
基本健康检查。

```bash
curl http://localhost:8082/health
```

**响应：**
```json
{
  "status": "ok"
}
```

#### GET /health/ready
就绪检查，包含提供商状态。

```bash
curl http://localhost:8082/health/ready
```

**响应：**
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

### 消息端点

#### POST /v1/messages
使用 Anthropic API 格式发送消息。

```bash
curl -X POST http://localhost:8082/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: 你的-api-key" \
  -d '{
    "model": "openai/gpt-4o",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "你好！"}
    ]
  }'
```

**请求体：**
| 字段 | 类型 | 必需 | 描述 |
|-------|------|------|-------------|
| `model` | string | 是 | 模型标识符（例如：`openai/gpt-4o`）|
| `max_tokens` | integer | 是 | 最大生成的 token 数 |
| `messages` | array | 是 | 消息对象数组 |
| `stream` | boolean | 否 | 启用流式传输（默认：false）|

**响应：**
```json
{
  "id": "msg_123",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "你好！今天我可以帮你什么？"
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

### 模型端点

#### GET /v1/models
列出所有可用模型。

```bash
curl http://localhost:8082/v1/models
```

**响应：**
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
<summary><strong>🔧 高级 API 用法</strong></summary>

### 流式响应

设置 `stream: true` 启用流式传输：

```bash
curl -X POST http://localhost:8082/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: 你的-api-key" \
  -d '{
    "model": "openai/gpt-4o",
    "max_tokens": 1024,
    "stream": true,
    "messages": [
      {"role": "user", "content": "你好！"}
    ]
  }'
```

响应将以服务器发送事件（SSE）格式发送。

### 错误响应

所有错误都遵循 Anthropic API 错误格式：

```json
{
  "type": "invalid_request_error",
  "error": {
    "type": "invalid_request_error",
    "message": "model field is required"
  }
}
```

### 速率限制

代理不实现速率限制。速率限制由上游提供商处理。

### 认证

代理支持两种认证模式：

1. **服务端**：API key 在 `config.toml` 中配置
2. **客户端**（绕过）：客户端通过 `X-API-Key` 请求头提供自己的 API key

在绕过模式下，`X-API-Key` 请求头会被转发给提供商。

</details>

---

## 🎯 示例

<details>
<summary><strong>📝 示例 1：多个提供商</strong></summary>

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
# 使用 OpenAI
curl -d '{"model": "openai/gpt-4o", ...}'

# 使用 Anthropic
curl -d '{"model": "anthropic/claude-3-5-sonnet-20241022", ...}'

# 使用 Ollama
curl -d '{"model": "ollama/llama3.2:7b", ...}'
```

</details>

<details>
<summary><strong>📝 示例 2：模型映射</strong></summary>

```toml
[mappings]
"haiku" = "ollama/llama3.2:1b"
"sonnet" = "ollama/llama3.2:3b"
"opus" = "ollama/llama3.2:7b"
"claude" = "anthropic/claude-3-5-sonnet-20241022"
"gpt" = "openai/gpt-4o"
```

```bash
# 简单名称
curl -d '{"model": "haiku", ...}'   # 使用 ollama/llama3.2:1b
curl -d '{"model": "sonnet", ...}'  # 使用 ollama/llama3.2:3b
curl -d '{"model": "claude", ...}'  # 使用 anthropic/claude-3-5-sonnet-20241022
```

</details>

<details>
<summary><strong>📝 示例 3：自定义 OpenAI 兼容 API</strong></summary>

```toml
[[providers]]
name = "deepseek"
type = "openai"
api_base_url = "https://api.deepseek.com/v1"
api_key = "bypass"  # 让客户端提供自己的 key
models = ["deepseek-chat", "deepseek-coder"]
```

```bash
curl -X POST http://localhost:8082/v1/messages \
  -H "x-api-key: 你的-deepseek-api-key" \
  -d '{"model": "deepseek/deepseek-chat", ...}'
```

</details>

<details>
<summary><strong>📝 示例 4：使用 Ollama 的本地 LLM</strong></summary>

```toml
[[providers]]
name = "local"
type = "openai"
api_base_url = "http://localhost:11434/v1"
api_key = "bypass"
models = ["llama3.2:1b", "llama3.2:3b", "llama3.2:7b"]

[mappings]
"快速" = "local/llama3.2:1b"
"平衡" = "local/llama3.2:3b"
"强力" = "local/llama3.2:7b"
```

```bash
# 使用本地 LLM
curl -d '{"model": "local/llama3.2:7b", ...}'
curl -d '{"model": "强力", ...}'  # 同上
```

</details>

---

## 🛠️ 开发

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/nerdneilsfield/llm-to-anthropic.git
cd llm-to-anthropic

# 构建
go build -o llm-to-anthropic .

# 运行测试
go test ./...

# 运行验证测试
./test_validation.sh
```

### 项目结构

```
llm-to-anthropic/
├── cmd/                # CLI 命令
│   ├── proxy/         # 代理命令
│   └── root.go       # 根命令
├── internal/          # 内部包
│   ├── config/       # 配置
│   ├── server/       # HTTP 服务器
│   └── ...          # 其他内部包
├── pkg/              # 公共包
│   ├── provider/      # 提供商客户端
│   ├── api/          # API 处理器
│   └── logger/       # 日志
├── config.toml       # 配置文件
├── README.md         # 英文文档
├── README_zh.md     # 中文文档（本文件）
└── main.go           # 入口点
```

### 贡献指南

1. Fork 仓库
2. 创建功能分支（`git checkout -b feature/amazing-feature`）
3. 提交你的更改（使用 conventional commits：`feat:`、`fix:`、`docs:` 等）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 创建 Pull Request

---

## 📄 许可证

MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

---

## 🔒 安全最佳实践

1. **永远不要提交 API key** 到仓库
2. **使用环境变量** 进行敏感配置
3. **设置正确的文件权限** 对 config.toml（`chmod 600`）
4. **在生产环境中使用 HTTPS**
5. **保持镜像更新** 以获取安全补丁
6. **定期审查依赖项** 查找漏洞
7. **在提供商级别使用速率限制**
8. **监控日志** 查找可疑活动
9. **在反向代理中实现认证**（如果需要）
10. **定期备份** 配置文件

---

## 🤝 支持

- 📖 [文档](CONFIGURATION_VALIDATION.md)
- 🐛 [问题跟踪](https://github.com/nerdneilsfield/llm-to-anthropic/issues)
- 💬 [讨论](https://github.com/nerdneilsfield/llm-to-anthropic/discussions)
- 📦 [Releases](https://github.com/nerdneilsfield/llm-to-anthropic/releases)
- 🐳 [Docker Hub](https://hub.docker.com/r/nerdneilsfield/llm-to-anthropic)
- 📦 [GitHub Container Registry](https://github.com/nerdneilsfield/llm-to-anthropic/pkgs/container/llm-to-anthropic)

---

<div align="center">

**由社区用 ❤️ 制作**

[⬆ 返回顶部](#-llm-到-anthropic-代理)

</div>

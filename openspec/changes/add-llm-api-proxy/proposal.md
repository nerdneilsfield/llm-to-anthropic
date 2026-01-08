# Change: Add LLM API Proxy

## Why
Enable users to use various LLM providers (OpenAI, Google Gemini, Anthropic) with a unified Anthropic-compatible API interface, similar to the Python-based claude-code-proxy project but implemented in Go.

## What Changes
- Implement an HTTP server using Fiber v2 framework
- Add support for OpenAI-compatible API translation to Anthropic format
- Add support for Google Gemini API translation to Anthropic format
- Add optional direct Anthropic API proxy mode
- Implement model mapping configuration (haiku/sonnet to provider-specific models)
- Support both streaming and non-streaming responses
- Add configuration via environment variables
- **BREAKING**: This is a new feature, no breaking changes to existing code

## Impact
- Affected specs: New capability `llm-api-proxy`
- Affected code:
  - New `pkg/api/proxy/` package for API translation logic
  - New `cmd/proxy/` command for running the proxy server
  - New `internal/server/` for Fiber v2 HTTP server setup
  - Updates to `go.mod` for new dependencies (fiber, environment variable management)
  - Updates to `Dockerfile` for proxy service support

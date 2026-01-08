## Context
This project aims to create a Go-based LLM API proxy that translates various LLM provider APIs (OpenAI, Google Gemini, Anthropic) into a unified Anthropic-compatible format. This enables clients like Claude Code to work with different LLM backends seamlessly.

The reference implementation is a Python-based project using FastAPI and LiteLLM. We will implement similar functionality in Go using Fiber v2 for the HTTP server.

## Goals / Non-Goals

### Goals
- Provide a high-performance, production-ready proxy server in Go
- Support OpenAI-compatible APIs (including Azure OpenAI)
- Support Google Gemini APIs (both API key and Vertex AI ADC)
- Support direct Anthropic API proxy mode
- Maintain 100% API compatibility with Anthropic's v1 messages API
- Support both streaming and non-streaming responses
- Provide flexible model mapping configuration
- Enable graceful shutdown and health monitoring
- Maintain clean separation of concerns (translation, backend clients, server)

### Non-Goals
- Support for other LLM providers beyond OpenAI and Google Gemini in the initial implementation
- Rate limiting or request queuing (can be added later)
- Request/response caching (can be added later)
- Authentication/authorization layers (delegated to reverse proxy)
- WebSocket support (Anthropic API uses SSE for streaming)
- Multi-tenancy or user-specific routing

## Decisions

### 1. Web Framework: Fiber v2
**Decision**: Use Fiber v2 for the HTTP server.

**Rationale**:
- High performance (built on Fasthttp, similar to Express.js)
- Excellent middleware ecosystem
- Simple and intuitive API
- Good Go community and maintenance
- Built-in support for streaming responses
- Lightweight and fast startup (important for containers)

**Alternatives considered**:
- Gin: Good but Fiber is generally faster due to Fasthttp
- Net/http: Too low-level, would require more boilerplate
- Echo: Good alternative but Fiber has better middleware ecosystem
- Chi: Lightweight but fewer built-in features

### 2. API Translation Architecture
**Decision**: Separate translation layer with distinct packages for each provider.

**Rationale**:
- Clear separation of concerns
- Easy to add new providers in the future
- Testable in isolation
- Each provider's quirks can be isolated to its own package
- Follows Go's interface-based design patterns

**Package structure**:
```
pkg/api/proxy/
  ├── anthropic/      # Anthropic API types
  ├── openai/         # OpenAI translation logic
  ├── gemini/         # Gemini translation logic
  └── translator.go   # Generic translation interface
```

### 3. Model Mapping Strategy
**Decision**: Configuration-based model mapping with prefix handling.

**Rationale**:
- Allows users to override defaults via environment variables
- Supports explicit model specification (e.g., `openai/gpt-4o`)
- Maintains compatibility with existing clients
- Prefix-based approach prevents conflicts

**Mapping logic**:
1. If model name includes a prefix (`openai/`, `gemini/`, `anthropic/`), route to that provider directly
2. If model name is `haiku` or `sonnet`, map based on `PREFERRED_PROVIDER`:
   - `openai` → `SMALL_MODEL` / `BIG_MODEL` with `openai/` prefix
   - `google` → `SMALL_MODEL` / `BIG_MODEL` with `gemini/` prefix
   - `anthropic` → Use actual Anthropic models with `anthropic/` prefix
3. If provider-specific model name provided without prefix, add prefix automatically

### 4. Configuration Management
**Decision**: Use a struct-based configuration with environment variable loading.

**Rationale**:
- Type-safe configuration
- Easy to validate
- Simple to add new configuration options
- Supports hot-reload if needed in the future

**Library choice**: Use `github.com/kelseyhightower/envconfig` or `github.com/spf13/viper`

### 5. Streaming Response Handling
**Decision**: Use Fiber's streaming support with Server-Sent Events (SSE) for streaming responses.

**Rationale**:
- Anthropic API uses SSE for streaming
- Fiber has built-in SSE support
- Maintains compatibility with Anthropic clients
- Allows real-time streaming to clients

**Implementation approach**:
- Use `c.Context().Stream()` for Fiber
- Parse streaming responses from backends
- Transform events on-the-fly
- Forward transformed events to client

### 6. Error Handling
**Decision**: Structured error types with HTTP status code mapping.

**Rationale**:
- Consistent error responses across providers
- Clear error messages for debugging
- Proper HTTP status codes for client compatibility

**Error types**:
- Validation errors (400)
- Authentication errors (401)
- Rate limit errors (429)
- Server errors (500)
- Timeout errors (504)

### 7. HTTP Client Strategy
**Decision**: Use `net/http` with custom wrapper for provider-specific logic.

**Rationale**:
- Go's standard library is sufficient
- Avoid unnecessary dependencies
- Custom wrapper allows for provider-specific headers/auth
- Easy to add retry logic and middleware

**Features**:
- Timeout configuration
- Retry with exponential backoff
- Request/response logging
- Metrics collection (optional)

### 8. Testing Strategy
**Decision**: Multi-layer testing approach.

**Rationale**:
- Unit tests for translation logic (fast, isolated)
- Integration tests for backend clients (requires API keys or mocks)
- End-to-end tests for the full proxy stack
- Contract tests to ensure API compatibility

**Testing tools**:
- Standard `testing` package
- `httptest` for HTTP server tests
- Provider-specific test fixtures for translation tests
- Environment-specific test configurations

### 9. Logging
**Decision**: Use existing zap logger from shlogin package.

**Rationale**:
- Already integrated in the project
- Structured logging
- Performance-oriented
- Supports log levels and sampling

**Log levels**:
- Debug: Detailed request/response bodies (only in verbose mode)
- Info: Request metadata, model routing, errors
- Warn: Retry attempts, degraded performance
- Error: Failed requests, client errors

### 10. Health Monitoring
**Decision**: Implement health check endpoint with backend connectivity checks.

**Rationale**:
- Enables orchestration systems (Kubernetes, Docker Compose)
- Helps detect misconfiguration early
- Provides visibility into backend availability

**Health checks**:
- `/health`: Basic server health (always 200)
- `/health/ready`: Backend connectivity checks
- `/metrics`: Prometheus metrics (optional, for future)

## Risks / Trade-offs

### Risk 1: API Compatibility
**Risk**: Anthropic API changes could break the proxy.

**Mitigation**:
- Pin to specific Anthropic API version
- Implement comprehensive contract tests
- Monitor for API changes
- Support versioning in the proxy itself

### Risk 2: Performance Overhead
**Risk**: Translation layer adds latency.

**Mitigation**:
- Minimize allocation in hot paths
- Use streaming where possible
- Benchmark performance regularly
- Optimize critical paths if needed

### Risk 3: Provider-Specific Quirks
**Risk**: Different providers have incompatible features or behaviors.

**Mitigation**:
- Document known limitations
- Provide clear error messages for unsupported features
- Gracefully degrade where possible
- Allow users to bypass translation with direct model access

### Risk 4: Streaming Complexity
**Risk**: Streaming responses are complex to implement correctly.

**Mitigation**:
- Thorough testing with real streaming responses
- Use proven patterns from reference implementation
- Implement proper error handling for interrupted streams
- Test with various client implementations

### Trade-off: Memory Usage vs. Simplicity
**Decision**: Prioritize simplicity over optimization initially.

**Rationale**:
- Go's garbage collector handles most memory efficiently
- Premature optimization is wasteful
- Can optimize based on profiling data later
- Simple code is easier to maintain

## Migration Plan

Since this is a new feature, no migration is needed for existing code. However:

1. **Development Phase**:
   - Start with OpenAI provider (simplest API)
   - Add Gemini provider
   - Add Anthropic direct proxy mode
   - Add streaming support last

2. **Testing Phase**:
   - Test with real API keys (configurable)
   - Test with Claude Code client
   - Load testing for performance validation

3. **Deployment Phase**:
   - Deploy with feature flag or separate binary
   - Monitor logs and metrics
   - Gather feedback before GA

4. **Rollback Plan**:
   - Keep previous version available
   - Simple binary rollback
   - No database state to worry about

## Open Questions

1. **Azure OpenAI Support**: Should we add Azure OpenAI in the initial implementation?
   - Recommendation: Defer to v2, as it requires additional authentication (Azure AD)

2. **Model Discovery**: Should we implement automatic model discovery from providers?
   - Recommendation: Start with hardcoded lists, add discovery later for flexibility

3. **Rate Limiting**: Should we implement client-side rate limiting?
   - Recommendation: Defer to reverse proxy or add in v2 if needed

4. **Request Queuing**: Should we implement request queuing for cost optimization?
   - Recommendation: Defer to v2, not in scope for MVP

5. **Metrics**: Should we add Prometheus metrics out of the box?
   - Recommendation: Add basic metrics (request count, latency, errors) in v1, expand in v2

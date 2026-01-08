# LLM API Proxy Specification

This specification defines the requirements for the LLM API proxy that translates various LLM provider APIs into a unified Anthropic-compatible format.

## ADDED Requirements

### Requirement: HTTP Server
The system SHALL provide an HTTP server using Fiber v2 framework that accepts Anthropic-compatible API requests and returns Anthropic-compatible responses.

#### Scenario: Server starts successfully
- **WHEN** the proxy server is started with valid configuration
- **THEN** the server listens on the configured host and port
- **AND** the health check endpoint at `/health` responds with HTTP 200

#### Scenario: Server starts with invalid configuration
- **WHEN** the proxy server is started with missing required environment variables
- **THEN** the server fails to start with a clear error message indicating which configuration is missing

#### Scenario: Server handles graceful shutdown
- **WHEN** the server receives SIGTERM or SIGINT signal
- **THEN** the server stops accepting new connections
- **AND** in-flight requests complete before shutdown
- **AND** the server exits cleanly

### Requirement: Anthropic API v1 Messages Endpoint
The system SHALL implement the Anthropic API v1 messages endpoint (`POST /v1/messages`) with full compatibility for streaming and non-streaming responses.

#### Scenario: Non-streaming message to OpenAI
- **WHEN** a client sends a POST request to `/v1/messages` with model `haiku` and `stream: false`
- **AND** `PREFERRED_PROVIDER` is set to `openai`
- **THEN** the request is translated to OpenAI API format
- **AND** the OpenAI API is called with the mapped model (e.g., `gpt-4.1-mini`)
- **AND** the response is translated back to Anthropic format
- **AND** the client receives a valid Anthropic-compatible response with HTTP 200

#### Scenario: Streaming message to Gemini
- **WHEN** a client sends a POST request to `/v1/messages` with model `sonnet` and `stream: true`
- **AND** `PREFERRED_PROVIDER` is set to `google`
- **THEN** the request is translated to Gemini API format
- **AND** the Gemini API is called with the mapped model (e.g., `gemini-2.5-pro`)
- **AND** streaming responses are translated to Anthropic SSE format in real-time
- **AND** the client receives a valid Anthropic-compatible SSE stream

#### Scenario: Direct model specification with provider prefix
- **WHEN** a client sends a POST request to `/v1/messages` with model `openai/gpt-4o`
- **THEN** the request is routed to OpenAI provider without additional model mapping
- **AND** the model name is passed directly to OpenAI API as `gpt-4o`
- **AND** the response is translated back to Anthropic format

#### Scenario: Invalid model name
- **WHEN** a client sends a POST request to `/v1/messages` with an unknown model name
- **THEN** the server responds with HTTP 400
- **AND** the error message includes the invalid model name
- **AND** the error message lists available models

### Requirement: Model Listing Endpoint
The system SHALL implement the models listing endpoint (`GET /v1/models`) that returns all available models from all configured providers.

#### Scenario: List all available models
- **WHEN** a client sends a GET request to `/v1/models`
- **THEN** the server returns a list of all configured models
- **AND** each model includes the Anthropic-compatible model ID
- **AND** OpenAI models are prefixed with `openai/`
- **AND** Gemini models are prefixed with `gemini/`
- **AND** Anthropic models are prefixed with `anthropic/`

#### Scenario: List models with no providers configured
- **WHEN** a client sends a GET request to `/v1/models` with no API keys configured
- **THEN** the server returns an empty list with HTTP 200
- **OR** the server returns an error with HTTP 503 indicating no providers are available

### Requirement: OpenAI Provider Support
The system SHALL support OpenAI-compatible API as a backend provider, including Azure OpenAI.

#### Scenario: OpenAI API key authentication
- **WHEN** `OPENAI_API_KEY` is configured in environment variables
- **THEN** requests to OpenAI are authenticated with the provided API key via `Authorization` header
- **AND** the header format is `Bearer sk-...`

#### Scenario: OpenAI streaming response translation
- **WHEN** OpenAI returns a streaming response in its format
- **THEN** each chunk is translated to Anthropic SSE format
- **AND** `delta.content` is mapped to `content_block.delta.text`
- **AND** `finish_reason` is preserved
- **AND** the SSE events follow Anthropic's format

#### Scenario: OpenAI error handling
- **WHEN** OpenAI returns an error (e.g., 401 Unauthorized, 429 Rate Limited)
- **THEN** the error is translated to Anthropic error format
- **AND** the HTTP status code is preserved
- **AND** the error message includes the original error from OpenAI

### Requirement: Google Gemini Provider Support
The system SHALL support Google Gemini API as a backend provider, with both API key and Vertex AI authentication methods.

#### Scenario: Gemini API key authentication
- **WHEN** `GEMINI_API_KEY` is configured and `USE_VERTEX_AUTH` is `false` or unset
- **THEN** requests to Gemini are authenticated with the provided API key via URL parameter or header
- **AND** the authentication format follows Google's API specification

#### Scenario: Gemini Vertex AI authentication
- **WHEN** `USE_VERTEX_AUTH` is set to `true`
- **AND** `VERTEX_PROJECT` and `VERTEX_LOCATION` are configured
- **THEN** requests to Gemini are authenticated using Application Default Credentials (ADC)
- **AND** the request is sent to the Vertex AI endpoint for the configured project and region

#### Scenario: Gemini streaming response translation
- **WHEN** Gemini returns a streaming response in its format
- **THEN** each chunk is translated to Anthropic SSE format
- **AND** the content blocks are mapped to Anthropic's content block format
- **AND** the role information is preserved
- **AND** the finish_reason is properly translated

#### Scenario: Gemini error handling
- **WHEN** Gemini returns an error (e.g., 400 Bad Request, 403 Permission Denied)
- **THEN** the error is translated to Anthropic error format
- **AND** the HTTP status code is mapped to an appropriate Anthropic status code
- **AND** the error message includes relevant details from Gemini's error response

### Requirement: Anthropic Direct Proxy Mode
The system SHALL support direct proxying to Anthropic API when `PREFERRED_PROVIDER` is set to `anthropic`.

#### Scenario: Direct Anthropic proxy
- **WHEN** `PREFERRED_PROVIDER` is set to `anthropic`
- **AND** `ANTHROPIC_API_KEY` is configured
- **THEN** requests are forwarded directly to Anthropic's API without translation
- **AND** the `Authorization` header is passed through
- **AND** the model name (e.g., `haiku`, `sonnet`) is preserved
- **AND** responses are returned as-is without translation

#### Scenario: Direct Anthropic proxy without API key
- **WHEN** `PREFERRED_PROVIDER` is set to `anthropic` but `ANTHROPIC_API_KEY` is not configured
- **THEN** the server fails to start with an error message indicating the missing API key

### Requirement: Model Mapping Configuration
The system SHALL support configurable model mapping from Anthropic model names (haiku/sonnet) to provider-specific models.

#### Scenario: Default OpenAI mapping
- **WHEN** `PREFERRED_PROVIDER` is set to `openai` (default)
- **AND** `BIG_MODEL` and `SMALL_MODEL` are not configured
- **AND** a request is made with model `sonnet`
- **THEN** the request is mapped to `openai/gpt-4.1`
- **AND** a request with model `haiku` is mapped to `openai/gpt-4.1-mini`

#### Scenario: Default Google mapping
- **WHEN** `PREFERRED_PROVIDER` is set to `google`
- **AND** `BIG_MODEL` and `SMALL_MODEL` are not configured
- **AND** a request is made with model `sonnet`
- **THEN** the request is mapped to `gemini/gemini-2.5-pro`
- **AND** a request with model `haiku` is mapped to `gemini/gemini-2.5-flash`

#### Scenario: Custom model mapping
- **WHEN** `BIG_MODEL` is set to `custom-model-pro` and `SMALL_MODEL` is set to `custom-model-mini`
- **AND** `PREFERRED_PROVIDER` is set to `openai`
- **AND** a request is made with model `sonnet`
- **THEN** the request is mapped to `openai/custom-model-pro`
- **AND** a request with model `haiku` is mapped to `openai/custom-model-mini`

#### Scenario: Model with explicit provider prefix bypasses mapping
- **WHEN** a request is made with model `openai/gpt-4o`
- **THEN** the model is passed directly to OpenAI
- **AND** no model mapping logic is applied

### Requirement: Configuration Management
The system SHALL load configuration from environment variables with validation and defaults.

#### Scenario: Required configuration missing
- **WHEN** the server starts without any API keys configured
- **THEN** the server logs a warning but continues
- **AND** the `/health/ready` endpoint returns HTTP 503
- **AND** API requests fail with an error indicating no providers are available

#### Scenario: Optional configuration uses defaults
- **WHEN** optional environment variables are not set (e.g., `SERVER_PORT`, `PREFERRED_PROVIDER`)
- **THEN** sensible defaults are used (e.g., port 8082, provider `openai`)

#### Scenario: Invalid configuration value
- **WHEN** an environment variable has an invalid value (e.g., `PREFERRED_PROVIDER=invalid`)
- **THEN** the server fails to start with a clear error message
- **AND** the error message indicates which configuration value is invalid

### Requirement: Request Logging
The system SHALL log all incoming requests and outgoing responses with sufficient detail for debugging and monitoring.

#### Scenario: Request logging in verbose mode
- **WHEN** the server is started with verbose logging enabled
- **THEN** each request logs the request method, path, headers, and body
- **AND** each response logs the status code, headers, and body
- **AND** model routing decisions are logged

#### Scenario: Request logging in normal mode
- **WHEN** the server is started without verbose logging
- **THEN** each request logs the request method, path, and model
- **AND** each response logs the status code and response time
- **AND** request/response bodies are not logged to protect privacy

### Requirement: Error Handling
The system SHALL handle errors gracefully and return Anthropic-compatible error responses.

#### Scenario: Backend API timeout
- **WHEN** a backend API (OpenAI, Gemini) times out
- **THEN** the proxy returns HTTP 504 Gateway Timeout
- **AND** the error message indicates a timeout occurred
- **AND** the original error is logged

#### Scenario: Backend API 5xx error
- **WHEN** a backend API returns a 5xx error
- **THEN** the proxy returns HTTP 502 Bad Gateway
- **AND** the error message indicates an upstream server error
- **AND** the original error details are logged

#### Scenario: Invalid request format
- **WHEN** a client sends a request with invalid JSON
- **THEN** the proxy returns HTTP 400 Bad Request
- **AND** the error message indicates invalid JSON
- **AND** the error includes the parsing error details

#### Scenario: Missing required fields
- **WHEN** a client sends a request missing required fields (e.g., `model`, `messages`)
- **THEN** the proxy returns HTTP 400 Bad Request
- **AND** the error message indicates which required field is missing
- **AND** the error message follows Anthropic's error format

### Requirement: CORS Support
The system SHALL support Cross-Origin Resource Sharing (CORS) to allow web-based clients to access the proxy.

#### Scenario: CORS headers returned
- **WHEN** a client sends an OPTIONS request with an `Origin` header
- **THEN** the server responds with appropriate CORS headers
- **AND** the `Access-Control-Allow-Origin` header matches the request origin or is set to `*`
- **AND** allowed methods are listed in `Access-Control-Allow-Methods`
- **AND** allowed headers are listed in `Access-Control-Allow-Headers`

#### Scenario: CORS preflight handling
- **WHEN** a client sends a preflight OPTIONS request
- **THEN** the server responds with HTTP 200
- **AND** appropriate CORS headers are included
- **AND** no actual request is made to the backend

### Requirement: Health Check Endpoints
The system SHALL provide health check endpoints for monitoring and orchestration.

#### Scenario: Basic health check
- **WHEN** a client sends a GET request to `/health`
- **THEN** the server responds with HTTP 200
- **AND** the response body includes `"status": "ok"`
- **AND** no backend connectivity checks are performed

#### Scenario: Readiness health check with healthy backends
- **WHEN** a client sends a GET request to `/health/ready`
- **AND** at least one backend provider is configured and reachable
- **THEN** the server responds with HTTP 200
- **AND** the response includes the status of each configured provider

#### Scenario: Readiness health check with unhealthy backends
- **WHEN** a client sends a GET request to `/health/ready`
- **AND** all configured backend providers are unreachable
- **THEN** the server responds with HTTP 503
- **AND** the response indicates which providers are unhealthy

### Requirement: Content Type Handling
The system SHALL correctly handle and set content types for all requests and responses.

#### Scenario: Request with correct content type
- **WHEN** a client sends a POST request with `Content-Type: application/json`
- **THEN** the request body is parsed as JSON
- **AND** the request is forwarded to the backend

#### Scenario: Request with incorrect content type
- **WHEN** a client sends a POST request without `Content-Type: application/json`
- **THEN** the server responds with HTTP 415 Unsupported Media Type
- **AND** the error message indicates correct content type is required

#### Scenario: Response content type for non-streaming
- **WHEN** responding to a non-streaming request
- **THEN** the `Content-Type` header is set to `application/json`

#### Scenario: Response content type for streaming
- **WHEN** responding to a streaming request
- **THEN** the `Content-Type` header is set to `text/event-stream`
- **AND** the `Cache-Control` header is set to `no-cache`
- **AND** the `Connection` header is set to `keep-alive`

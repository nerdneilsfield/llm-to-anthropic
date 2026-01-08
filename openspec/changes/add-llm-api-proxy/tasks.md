## 1. Project Setup & Dependencies
- [ ] 1.1 Add Fiber v2 dependency to go.mod
- [ ] 1.2 Add environment variable management (e.g., godotenv or similar)
- [ ] 1.3 Add HTTP client libraries for OpenAI and Google Gemini APIs
- [ ] 1.4 Update Makefile/justfile with new build targets for proxy

## 2. Configuration Management
- [ ] 2.1 Create configuration struct for environment variables
- [ ] 2.2 Implement configuration loading with validation
- [ ] 2.3 Add support for:
  - ANTHROPIC_API_KEY
  - OPENAI_API_KEY
  - GEMINI_API_KEY
  - VERTEX_PROJECT
  - VERTEX_LOCATION
  - USE_VERTEX_AUTH
  - PREFERRED_PROVIDER (openai/google/anthropic)
  - BIG_MODEL
  - SMALL_MODEL
  - SERVER_PORT
  - SERVER_HOST
- [ ] 2.4 Create .env.example file with all configuration options

## 3. API Translation Layer
- [ ] 3.1 Define Anthropic API request/response structures
- [ ] 3.2 Define OpenAI API request/response structures
- [ ] 3.3 Define Google Gemini API request/response structures
- [ ] 3.4 Implement Anthropic to OpenAI request translator
- [ ] 3.5 Implement OpenAI to Anthropic response translator
- [ ] 3.6 Implement Anthropic to Gemini request translator
- [ ] 3.7 Implement Gemini to Anthropic response translator
- [ ] 3.8 Implement model mapping logic (haiku/sonnet → provider models)

## 4. Backend Clients
- [ ] 4.1 Create OpenAI client with API key authentication
- [ ] 4.2 Create Gemini client with API key authentication
- [ ] 4.3 Create Gemini client with Vertex AI ADC authentication
- [ ] 4.4 Create Anthropic client for direct proxy mode
- [ ] 4.5 Implement streaming response handling for all clients
- [ ] 4.6 Implement error handling and retry logic

## 5. HTTP Server (Fiber v2)
- [ ] 5.1 Set up Fiber v2 server with health check endpoint
- [ ] 5.2 Implement Anthropic API v1 messages endpoint (`POST /v1/messages`)
- [ ] 5.3 Implement model listing endpoint (`GET /v1/models`)
- [ ] 5.4 Add middleware for request logging
- [ ] 5.5 Add middleware for error handling
- [ ] 5.6 Add CORS support
- [ ] 5.7 Implement graceful shutdown handling

## 6. Model Management
- [ ] 6.1 Define supported OpenAI models list
- [ ] 6.2 Define supported Gemini models list
- [ ] 6.3 Implement model prefix handling (openai/, gemini/, anthropic/)
- [ ] 6.4 Add model validation logic
- [ ] 6.5 Implement model discovery/registration

## 7. CLI Integration
- [ ] 7.1 Create `cmd/proxy` package
- [ ] 7.2 Add `proxy` subcommand to root Cobra command
- [ ] 7.3 Add flags for configuration override
- [ ] 7.4 Add verbose logging support

## 8. Testing
- [ ] 8.1 Write unit tests for API translation layer
- [ ] 8.2 Write unit tests for model mapping logic
- [ ] 8.3 Write integration tests for OpenAI backend
- [ ] 8.4 Write integration tests for Gemini backend
- [ ] 8.5 Write integration tests for Anthropic backend
- [ ] 8.6 Write end-to-end tests for proxy server
- [ ] 8.7 Add test coverage reporting

## 9. Documentation
- [ ] 9.1 Update README.md with proxy usage instructions
- [ ] 9.2 Add API documentation
- [ ] 9.3 Add configuration guide
- [ ] 9.4 Add model mapping documentation
- [ ] 9.5 Add Docker setup instructions

## 10. Deployment
- [ ] 10.1 Update Dockerfile for proxy service
- [ ] 10.2 Update GoReleaser configuration
- [ ] 10.3 Add health check configuration
- [ ] 10.4 Test production deployment

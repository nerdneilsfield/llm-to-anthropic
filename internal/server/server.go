package server

import (
	openai "github.com/nerdneilsfield/llm-to-anthropic/pkg/provider/openai"
	anthropic_provider "github.com/nerdneilsfield/llm-to-anthropic/pkg/provider/anthropic"
	gemini "github.com/nerdneilsfield/llm-to-anthropic/pkg/provider/gemini"
	translators "github.com/nerdneilsfield/llm-to-anthropic/pkg/api/proxy/translators"
	"fmt"
	"time"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/nerdneilsfield/llm-to-anthropic/internal/config"
	"github.com/nerdneilsfield/llm-to-anthropic/pkg/api/proxy/anthropic"
	"github.com/nerdneilsfield/llm-to-anthropic/pkg/api/proxy"
	"go.uber.org/zap"
)

// Server wraps the Fiber HTTP server
type Server struct {
	app           *fiber.App
	cfg           *config.Config
	modelManager  *proxy.ModelManager
	logger        *zap.Logger
}


// customErrorHandler is a custom error handler
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(anthropic.ErrorResponse{
		Type: "internal_error",
		Error: &anthropic.Error{
			Type:    "internal_error",
			Message: err.Error(),
		},
	})
}
// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, logger *zap.Logger) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "llm-api-proxy",
		ServerHeader:  "llm-api-proxy",
		ReadTimeout:   time.Duration(cfg.GetReadTimeout()) * time.Second,
		WriteTimeout:  time.Duration(cfg.GetWriteTimeout()) * time.Second,
		IdleTimeout:   120 * time.Second,
		ErrorHandler:  customErrorHandler,
	})

	// Add middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-API-Key",
		ExposeHeaders:    "Content-Type",
		AllowCredentials: false,
		MaxAge:          86400,
	}))

	return &Server{
		app:          app,
		cfg:          cfg,
		modelManager:  proxy.NewModelManager(cfg),
		logger:       logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Register routes
	s.registerRoutes()

	// Start server
	addr := fmt.Sprintf("%s:%d", s.cfg.GetHost(), s.cfg.GetPort())
	s.logger.Info("Starting server", zap.String("address", addr))
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	s.logger.Info("Shutting down server")
	return s.app.Shutdown()
}

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {
	// Health check endpoints
	s.app.Get("/health", s.handleHealth)
	s.app.Get("/health/ready", s.handleReady)

	// Anthropic API v1 endpoints
	api := s.app.Group("/v1")
	api.Post("/messages", s.handleMessages)
	api.Get("/models", s.handleModels)
}

// handleHealth handles the basic health check endpoint
func (s *Server) handleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

// handleReady handles the readiness health check endpoint
func (s *Server) handleReady(c *fiber.Ctx) error {
	status := fiber.Map{
		"status": "ready",
	}

	// Check provider status
	providers := fiber.Map{}

	for _, provider := range s.cfg.Providers {
		if provider.ParsedAPIKey != "" || provider.IsBypass {
			providers[provider.Name] = "configured"
		} else {
			providers[provider.Name] = "not_configured"
		}
	}

	status["providers"] = providers
	status["total_providers"] = len(s.cfg.Providers)
	status["total_mappings"] = len(s.cfg.Mappings)

	return c.JSON(status)
}

// handleMessages handles the Anthropic v1 messages endpoint
func (s *Server) handleMessages(c *fiber.Ctx) error {
	// Extract API key from request header (supports both formats)
	apiKey := c.Get("X-Api-Key")
	if apiKey == "" {
		apiKey = c.Get("x-api-key")
	}

	// Parse request
	var req anthropic.MessageRequest
	if err := c.BodyParser(&req); err != nil {
		s.logger.Error("Failed to parse request", zap.Error(err))
		return c.Status(400).JSON(anthropic.ErrorResponse{
			Type: "invalid_request_error",
			Error: &anthropic.Error{
				Type:    "invalid_request_error",
				Message: fmt.Sprintf("Invalid JSON: %v", err),
			},
		})
	}

	// Validate request
	if req.Model == "" {
		return c.Status(400).JSON(anthropic.ErrorResponse{
			Type: "invalid_request_error",
			Error: &anthropic.Error{
				Type:    "invalid_request_error",
				Message: "model field is required",
			},
		})
	}

	if req.MaxTokens <= 0 {
		return c.Status(400).JSON(anthropic.ErrorResponse{
			Type: "invalid_request_error",
			Error: &anthropic.Error{
				Type:    "invalid_request_error",
				Message: "max_tokens must be greater than 0",
			},
		})
	}

	if len(req.Messages) == 0 {
		return c.Status(400).JSON(anthropic.ErrorResponse{
			Type: "invalid_request_error",
			Error: &anthropic.Error{
				Type:    "invalid_request_error",
				Message: "messages field is required and must be non-empty",
			},
		})
	}

	// Parse model to determine provider
	model, err := s.modelManager.ParseModel(req.Model)
	if err != nil {
		s.logger.Error("Failed to parse model", zap.String("model", req.Model), zap.Error(err))
		return c.Status(400).JSON(anthropic.ErrorResponse{
			Type: "invalid_request_error",
			Error: &anthropic.Error{
				Type:    "invalid_request_error",
				Message: fmt.Sprintf("Invalid model: %v", err),
			},
		})
	}

	// Log request (don't log API key)
	s.logger.Info("Handling message request",
		zap.String("model", req.Model),
		zap.String("provider", model.Provider.Name),
		zap.Bool("stream", req.Stream),
		zap.Bool("has_api_key", apiKey != ""),
	)

	// Handle streaming vs non-streaming
	if req.Stream {
		return s.handleStreamingMessage(c, &req, model, apiKey)
	}

	return s.handleNonStreamingMessage(c, &req, model, apiKey)
}

// handleNonStreamingMessage handles non-streaming message requests
func (s *Server) handleNonStreamingMessage(c *fiber.Ctx, req *anthropic.MessageRequest, model *proxy.Model, apiKey string) error {
	// Translate request to provider format
	providerReq, err := s.translateRequest(req, model)
	if err != nil {
		s.logger.Error("Failed to translate request", zap.Error(err))
		return c.Status(500).JSON(anthropic.ErrorResponse{
			Type: "internal_error",
			Error: &anthropic.Error{
				Type:    "internal_error",
				Message: "Failed to translate request",
			},
		})
	}

	// Send request to provider with API key
	resp, err := s.sendToProvider(model, providerReq, apiKey)
	if err != nil {
		s.logger.Error("Provider request failed", zap.Error(err))
		return s.handleProviderError(c, err)
	}

	// Translate response back to Anthropic format
	anthropicResp, err := s.translateResponse(model, resp)
	if err != nil {
		s.logger.Error("Failed to translate response", zap.Error(err))
		return c.Status(500).JSON(anthropic.ErrorResponse{
			Type: "internal_error",
			Error: &anthropic.Error{
				Type:    "internal_error",
				Message: "Failed to translate response",
			},
		})
	}

	return c.JSON(anthropicResp)
}

// handleStreamingMessage handles streaming message requests
func (s *Server) handleStreamingMessage(c *fiber.Ctx, req *anthropic.MessageRequest, model *proxy.Model, apiKey string) error {
	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	// Translate request to provider format
	providerReq, err := s.translateRequest(req, model)
	if err != nil {
		s.logger.Error("Failed to translate request", zap.Error(err))
		return s.writeStreamError(c, err)
	}

	// Send streaming request to provider with API key
	stream, err := s.sendStreamToProvider(model, providerReq, apiKey)
	if err != nil {
		s.logger.Error("Provider stream request failed", zap.Error(err))
		return s.writeStreamError(c, err)
	}
	defer stream.Close()

	// Translate streaming response back to Anthropic SSE format
	if err := s.translateStream(model, stream, c); err != nil {
		s.logger.Error("Failed to translate stream", zap.Error(err))
		return err
	}

	return nil
}

// handleModels handles the models listing endpoint

// writeStreamError writes an error to the stream
func (s *Server) writeStreamError(c *fiber.Ctx, err error) error {
	fmt.Fprintf(c, "event: error\ndata: %s\n\n", err.Error())
	return nil
}
func (s *Server) handleModels(c *fiber.Ctx) error {
	models := s.modelManager.GetAvailableModels()
	return c.JSON(anthropic.ModelsResponse{
		Data: convertModelsToAnthropic(models),
	})
}



// convertModelsToAnthropic converts proxy models to Anthropic format
func convertModelsToAnthropic(models []proxy.Model) []anthropic.Model {
	anthropicModels := make([]anthropic.Model, 0, len(models))
	
	for _, model := range models {
		anthropicModels = append(anthropicModels, anthropic.Model{
			ID:        model.ID,
			Name:      model.Name,
			MaxTokens: 0, // TODO: Get from provider if available
			Type:      "model",
			Display:    model.Name,
			CreatedAt:  "",
		})
	}
	
	return anthropicModels
}
// Helper methods - implemented with provider clients
func (s *Server) translateRequest(req *anthropic.MessageRequest, model *proxy.Model) (interface{}, error) {
	switch model.Provider.Type {
	case "openai":
		return translators.TranslateAnthropicToOpenAI(req, model.Name)
	case "anthropic":
		return translators.TranslateAnthropicToAnthropic(req)
	case "gemini":
		return translators.TranslateAnthropicToGemini(req, model.Name)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", model.Provider.Type)
	}
}

func (s *Server) sendToProvider(model *proxy.Model, req interface{}, apiKey string) ([]byte, error) {
	client := s.getProviderClient(model.Provider)
	
	if apiKey != "" {
		return client.SendRequest(model.Name, req, apiKey)
	}
	return client.SendRequest(model.Name, req)
}

func (s *Server) sendStreamToProvider(model *proxy.Model, req interface{}, apiKey string) (io.ReadCloser, error) {
	client := s.getProviderClient(model.Provider)
	
	if apiKey != "" {
		return client.SendStream(model.Name, req, apiKey)
	}
	return client.SendStream(model.Name, req)
}

func (s *Server) translateResponse(model *proxy.Model, resp []byte) (*anthropic.MessageResponse, error) {
	switch model.Provider.Type {
	case "openai":
		return translators.TranslateOpenAIToAnthropic(resp)
	case "anthropic":
		return translators.TranslateAnthropicToAnthropicResponse(resp)
	case "gemini":
		return translators.TranslateGeminiToAnthropic(resp)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", model.Provider.Type)
	}
}


func (s *Server) getProviderClient(provider *config.Provider) proxy.ProviderClient {
	switch provider.Type {
	case "openai":
		return openai.NewClient(provider)
	case "anthropic":
		return anthropic_provider.NewClient(provider)
	case "gemini":
		return gemini.NewClient(provider)
	default:
		return openai.NewClient(provider)
	}
}
func (s *Server) translateStream(model *proxy.Model, stream io.Reader, w io.Writer) error {
	switch model.Provider.Type {
	case "openai":
		return translators.TranslateOpenAIStreamToAnthropicSSE(stream, w)
	case "anthropic":
		return translators.TranslateAnthropicStreamToAnthropicSSE(stream, w)
	case "gemini":
		return translators.TranslateGeminiStreamToAnthropicSSE(stream, w)
	default:
		return fmt.Errorf("unsupported provider type: %s", model.Provider.Type)
	}
}

func (s *Server) handleProviderError(c *fiber.Ctx, err error) error {
	return c.Status(500).JSON(anthropic.ErrorResponse{
		Type: "internal_error",
		Error: &anthropic.Error{
			Type:    "internal_error",
			Message: err.Error(),
		},
	})
}

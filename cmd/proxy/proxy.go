package proxy

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/nerdneilsfield/llm-to-anthropic/internal/config"
	"github.com/nerdneilsfield/llm-to-anthropic/internal/server"
	loggerPkg "github.com/nerdneilsfield/llm-to-anthropic/pkg/logger"
	"go.uber.org/zap"
)

// Cmd represents the proxy command (deprecated, use serve)
var Cmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start LLM API proxy server (deprecated: use 'serve' instead)",
	Long:  `Start a proxy server that translates various LLM provider APIs (OpenAI, Google Gemini, Anthropic) into a unified Anthropic-compatible format.

This command is deprecated. Please use 'serve' instead.`,
	Run:   runProxy,
	Hidden: false,
}

// NewServeCmd creates a new serve command
func NewServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start LLM API proxy server",
		Long:  `Start a proxy server that translates various LLM provider APIs (OpenAI, Google Gemini, Anthropic) into a unified Anthropic-compatible format.`,
		Run:   runProxy,
	}
}

// NewProxyCmd creates a new proxy command (alias for backward compatibility)
func NewProxyCmd() *cobra.Command {
	return Cmd
}

var (
	verbose bool
)

func init() {
	Cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}


// Get config path from args or use default
func getConfigPath(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}
func runProxy(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load(getConfigPath(args))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := loggerPkg.GetLogger(verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Log configuration
	logger.Info("Starting LLM API proxy",
		zap.Int("port", cfg.GetPort()),
		zap.Int("providers", len(cfg.Providers)),
		zap.Bool("verbose", verbose),
	)

	if verbose {
		logger.Debug("Configuration loaded",
			zap.String("host", cfg.GetHost()),
			zap.Int("port", cfg.GetPort()),
			zap.Int("mappings", len(cfg.Mappings)),
		)

		// Log provider status
		for _, provider := range cfg.Providers {
			logger.Debug("Provider configuration",
				zap.String("name", provider.Name),
				zap.String("type", provider.Type),
				zap.String("base_url", provider.BaseURL),
				zap.Bool("bypass", provider.IsBypass),
				zap.Int("models", len(provider.Models)),
			zap.Bool("has_api_key", provider.ParsedAPIKey != ""),
			)
		}
	}

	// Create server
	srv := server.NewServer(cfg, logger)

	// Setup graceful shutdown
	go setupSignalHandler(srv, logger)

	// Start server
	if err := srv.Start(); err != nil {
		logger.Error("Failed to start server", zap.Error(err))
		os.Exit(1)
	}
}

// setupSignalHandler sets up signal handling for graceful shutdown
func setupSignalHandler(srv *server.Server, logger *zap.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))

	if err := srv.Shutdown(); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Server shutdown complete")
	os.Exit(0)
}

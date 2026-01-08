package proxy

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/nerdneilsfield/go-template/internal/config"
	"github.com/nerdneilsfield/go-template/internal/server"
	loggerPkg "github.com/nerdneilsfield/go-template/pkg/logger"
	"go.uber.org/zap"
)

// Cmd represents the proxy command
var Cmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start LLM API proxy server",
	Long:  `Start a proxy server that translates various LLM provider APIs (OpenAI, Google Gemini, Anthropic) into a unified Anthropic-compatible format.`,
	Run:   runProxy,
}

var (
	cfgFile string
	verbose bool
)

func init() {
	Cmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file (default is .env in current directory)")
	Cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func runProxy(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Override with command line flags
	if verbose {
		cfg.Verbose = true
	}

	// Initialize logger
	logger, err := loggerPkg.GetLogger(cfg.Verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Log configuration
	logger.Info("Starting LLM API proxy",
		zap.String("provider", string(cfg.PreferredProvider)),
		zap.Int("port", cfg.ServerPort),
		zap.Bool("verbose", cfg.Verbose),
	)

	if cfg.Verbose {
		logger.Debug("Configuration loaded",
			zap.String("preferred_provider", string(cfg.PreferredProvider)),
			zap.String("big_model", cfg.BigModel),
			zap.String("small_model", cfg.SmallModel),
			zap.Bool("use_vertex_auth", cfg.UseVertexAuth),
			zap.String("vertex_project", cfg.VertexProject),
			zap.String("vertex_location", cfg.VertexLocation),
			zap.Bool("openai_configured", cfg.OpenAIKey != ""),
			zap.Bool("gemini_configured", cfg.GeminiAPIKey != ""),
			zap.Bool("anthropic_configured", cfg.AnthropicAPIKey != ""),
		)
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

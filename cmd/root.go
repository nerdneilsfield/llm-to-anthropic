package cmd

import (
	"fmt"

	loggerPkg "github.com/nerdneilsfield/shlogin/pkg/logger"
	"github.com/nerdneilsfield/llm-to-anthropic/cmd/proxy"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	verbose bool
	logger  = loggerPkg.GetLogger()
)

func newRootCmd(version string, buildTime string, gitCommit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llm-to-anthropic",
		Short: "LLM API proxy with Anthropic compatibility",
		Long: `LLM API proxy server that translates various LLM provider APIs 
(OpenAI, Google Gemini, Anthropic) into a unified Anthropic-compatible format.

Supports both server-side and client-side API key authentication.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				logger.SetVerbose(true)
			} else {
				logger.SetVerbose(false)
			}
			logger.Reset()
		},
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Add subcommands
	cmd.AddCommand(newVersionCmd(version, buildTime, gitCommit))
	cmd.AddCommand(proxy.NewServeCmd())
	cmd.AddCommand(proxy.NewProxyCmd()) // Alias for backward compatibility

	return cmd
}

func Execute(version string, buildTime string, gitCommit string) error {
	if err := newRootCmd(version, buildTime, gitCommit).Execute(); err != nil {
		logger.Fatal("error executing root command: %w", zap.Error(err))
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}

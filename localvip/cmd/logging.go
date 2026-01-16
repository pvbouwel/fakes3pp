package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func setupLogging(cmd *cobra.Command, args []string) {
	lvl := slog.LevelInfo
	isDebug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		panic("Coding error debug flag not defined")
	}
	if isDebug {
		lvl = slog.LevelDebug
	}
	options := slog.HandlerOptions{
		Level: lvl,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &options))
	slog.SetDefault(logger)
}

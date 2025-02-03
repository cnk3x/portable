package flags

import (
	"log/slog"
	"os"
	"path/filepath"
)

var rootCmd = &Command{Use: filepath.Base(os.Args[0])}

func init() {
	rootCmd = &Command{Use: filepath.Base(os.Args[0])}
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

// Execute execute the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error", "err", err)
		os.Exit(1)
	}
}

// RootSet set the root command
func RootSet(options ...Option) {
	for _, option := range options {
		option(rootCmd)
	}
}

// AddCommand add a command to the root command
func AddCommand(use string, options ...Option) *Command {
	c := &Command{Use: use}
	for _, option := range options {
		option(c)
	}
	rootCmd.AddCommand(c)
	return c
}

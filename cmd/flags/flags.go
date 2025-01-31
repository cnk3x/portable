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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error", "err", err)
		os.Exit(1)
	}
}

func RootSet(options ...Option) {
	for _, option := range options {
		option(rootCmd)
	}
}

func AddCommand(use string, options ...Option) *Command {
	c := &Command{Use: use}
	for _, option := range options {
		option(c)
	}
	rootCmd.AddCommand(c)
	return c
}

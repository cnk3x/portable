package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/cnk3x/portable/cmd/flags"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func init() {
	var verbose bool
	var debug bool
	flags.RootSet(flags.PersistentFlags(func(f *flags.FlagSet) {
		f.BoolVarP(&verbose, "verbose", "v", false, "verbose output")
		f.BoolVarP(&debug, "debug", "d", false, "debug output")
	}))
	flags.RootSet(flags.PersistentPreRun(func() {
		l := slog.LevelInfo
		if verbose {
			l = slog.LevelDebug
		}
		slog.SetDefault(slog.New(
			tint.NewHandler(os.Stderr, &tint.Options{
				Level:      l,
				TimeFormat: time.Kitchen,
				NoColor:    !isatty.IsTerminal(os.Stderr.Fd()),
				AddSource:  debug,
			}),
		))
	}))
}

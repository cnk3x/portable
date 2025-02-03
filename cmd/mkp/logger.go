package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/cnk3x/cl"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func init() {
	var verbose bool
	var debug bool
	cl.RootSet(cl.PersistentFlags(func(f *cl.FlagSet) {
		f.BoolVarP(&verbose, "verbose", "v", false, "verbose output")
		f.BoolVarP(&debug, "debug", "d", false, "debug output")
	}))
	cl.RootSet(cl.PersistentPreRun(func() {
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

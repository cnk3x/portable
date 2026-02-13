package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/cnk3x/cl"
	"github.com/cnk3x/portable"
	"github.com/lmittmann/tint"

	"github.com/mattn/go-isatty"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
			NoColor:    !isatty.IsTerminal(os.Stderr.Fd()),
			AddSource:  false,
		}),
	))

	cl.RootSet(cl.Description("manage portable app"))

	var (
		dirty bool
		all   bool
		depth int = 3
	)

	commandFlag := cl.Flags(
		cl.Val(&all, "all", "a", "install all apps in the current directory"),
		cl.Val(&depth, "depth", "", "depth of directory"),
		cl.Val(&dirty, "dirty", "", "dirty run"),
	)

	loadApps := func(args []string, all bool, depth int) []*portable.PortableApp {
		if len(args) == 0 {
			args = append(args, ".")
		}

		if !all {
			depth = 0
		}

		return portable.FindApps(args, depth)
	}

	cl.AddCommand(
		"install",
		cl.Aliases("i", "add"),
		cl.Description("install portable app"),
		commandFlag,
		cl.Run(func(c *cobra.Command, args []string) (err error) {
			for _, app := range loadApps(args, all, depth) {
				app.Install(dirty)
			}
			return
		}),
	)

	cl.AddCommand(
		"uninstall",
		cl.Aliases("un", "del", "rm", "remove"),
		cl.Description("uninstall portable app"),
		commandFlag,
		cl.Run(func(c *cobra.Command, args []string) {
			for _, app := range loadApps(args, all, depth) {
				app.Uninstall(dirty)
			}
		}),
	)

	cl.AddCommand(
		"list",
		cl.Aliases("ls"),
		cl.Description("list portable app"),
		commandFlag,
		cl.Run(func(c *cobra.Command, args []string) {
			for _, app := range loadApps(args, all, depth) {
				slog.Info(runewidth.FillRight(app.Name, 16), "path", app.ConfigPath())
			}
		}),
	)

	cl.Execute()
}

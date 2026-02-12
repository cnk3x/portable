package main

import (
	"log/slog"

	"github.com/cnk3x/cl"
	"github.com/cnk3x/portable"

	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

func main() {
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

	loadApps := func(args []string, all bool, depth int) (apps []*portable.PortableApp) {
		if len(args) == 0 {
			args = append(args, ".")
		}
		if !all {
			depth = 0
		}
		for _, a := range args {
			apps = append(apps, portable.LoadApps(a, depth)...)
		}
		return
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

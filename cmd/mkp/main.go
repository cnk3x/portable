package main

import (
	"cmp"
	"log/slog"

	"github.com/cnk3x/cl"
	"github.com/cnk3x/portable"

	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

func main() {
	cl.RootSet(cl.Description("manage portable app"))

	var (
		force bool
		all   bool
		depth int = 3
	)

	commandFlag := cl.Flags(
		cl.Val(&force, "force", "f", "if true, force to del the bind target, if false, only del the symlink"),
		cl.Val(&all, "all", "a", "install all apps in the current directory"),
		cl.Val(&depth, "depth", "d", "depth of directory"),
	)

	cl.AddCommand(
		"install",
		cl.Aliases("i", "add"),
		cl.Description("install portable app"),
		commandFlag,
		cl.Run(func(c *cobra.Command, apps []string) {
			if all {
				apps = portable.FindDirs(cmp.Or(cmp.Or(apps...), "."), depth)
			} else if len(apps) == 0 {
				apps = append(apps, ".")
			}

			for _, arg := range apps {
				slog.Info("install ", "path", arg)
				app, err := portable.LoadApp(arg)
				if err == nil {
					err = app.Install(force)
				}
				if err != nil {
					slog.Error("install ", "err", err)
				}
			}
		}),
	)

	cl.AddCommand(
		"uninstall",
		cl.Aliases("un", "del", "rm", "remove"),
		cl.Description("uninstall portable app"),
		commandFlag,
		cl.Run(func(c *cobra.Command, apps []string) {
			if all {
				apps = portable.FindDirs(cmp.Or(cmp.Or(apps...), "."), depth)
			} else if len(apps) == 0 {
				apps = append(apps, ".")
			}

			for _, arg := range apps {
				slog.Info("uninstall", "path", arg)
				app, err := portable.LoadApp(arg)
				if err == nil {
					err = app.Uninstall(force)
				}
				if err != nil {
					slog.Error("uninstall", "err", err)
				}
			}
		}),
	)

	cl.AddCommand(
		"list",
		cl.Aliases("ls"),
		cl.Description("list portable app"),
		commandFlag,
		cl.Run(func(c *cobra.Command, args []string) {
			args = portable.FindDirs(cmp.Or(cmp.Or(args...), "."), depth)
			for _, arg := range args {
				app, err := portable.LoadApp(arg)
				if err != nil {
					slog.Error(arg, "err", err)
				} else {
					slog.Info(runewidth.FillRight(app.Name, 16), "path", app.ConfigPath())
				}
			}
		}),
	)

	cl.Execute()
}

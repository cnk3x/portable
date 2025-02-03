package main

import (
	"cmp"
	"log/slog"

	"github.com/cnk3x/portable"
	"github.com/cnk3x/portable/cmd/flags"

	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

func main() {
	flags.RootSet(flags.Description("manage portable app"))

	var (
		force bool
		all   bool
		depth int = 3
	)

	commandFlag := flags.Flags(
		flags.Val(&force, "force", "f", "if true, force to del the bind target, if false, only del the symlink"),
		flags.Val(&all, "all", "a", "install all apps in the current directory"),
		flags.Val(&depth, "depth", "d", "depth of directory"),
	)

	flags.AddCommand(
		"install",
		flags.Aliases("i", "add"),
		flags.Description("install portable app"),
		commandFlag,
		flags.Run(func(c *cobra.Command, apps []string) {
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

	flags.AddCommand(
		"uninstall",
		flags.Aliases("un", "del", "rm", "remove"),
		flags.Description("uninstall portable app"),
		commandFlag,
		flags.Run(func(c *cobra.Command, apps []string) {
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

	flags.AddCommand(
		"list",
		flags.Aliases("ls"),
		flags.Description("list portable app"),
		commandFlag,
		flags.Run(func(c *cobra.Command, args []string) {
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

	flags.Execute()
}

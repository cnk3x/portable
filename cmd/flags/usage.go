package flags

func init() {
	rootCmd.SetUsageTemplate(usageTemplate)
}

const _ = usageTemplate

const usageTemplate = `Usage:
{{- if .Runnable}}
  {{.UseLine}}
{{- end}}
{{- if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]
{{- end}}
{{- if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}
{{- end}}
{{- if .HasExample}}

Examples:
{{.Example}}
{{- end}}
{{- if .HasAvailableSubCommands}}
{{- $cmds := .Commands}}
{{- if eq (len .Groups) 0}}

Available Commands:
{{- range $cmds}}
{{- if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}
{{- else}}
{{- range $group := .Groups}}

{{.Title}}
{{- range $cmds}}
{{- if (and (eq .GroupID $group.ID) .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}
{{- end}}
{{- if not .AllChildCommandsHaveGroup}}

Additional Commands:
{{- range $cmds}}
{{- if (and (eq .GroupID "") .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}

{{- if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end}}
{{- if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end}}
{{- if .HasHelpSubCommands}}

Additional help topics:
{{- range .Commands}}
{{- if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}
{{- end}}
{{- end}}
{{- end}}
{{- if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
{{- end}}
`

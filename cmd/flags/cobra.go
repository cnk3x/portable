package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	// FlagSet flag set
	FlagSet = pflag.FlagSet
	// Command command
	Command = cobra.Command
)

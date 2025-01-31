package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	FlagSet = pflag.FlagSet
	Command = cobra.Command
)

package flags

import "github.com/spf13/cobra"

// Expected arguments
func PositionalArgs(fn func(cmd *Command, args []string) error) Option {
	return func(c *Command) { c.Args = fn }
}

// RangeArgs returns an error if the number of args is not within the expected range.
func RangeArgs(min, max int) Option {
	return func(c *Command) { c.Args = cobra.RangeArgs(min, max) }
}

// MaximumNArgs returns an error if there are more than N args.
func MaximumNArgs(n int) Option {
	return func(c *Command) { c.Args = cobra.MaximumNArgs(n) }
}

// MinimumNArgs returns an error if there is not at least N args.
func MinimumNArgs(n int) Option {
	return func(c *Command) { c.Args = cobra.MinimumNArgs(n) }
}

// ExactArgs returns an error if there are not exactly n args.
func ExactArgs(n int) Option {
	return func(c *Command) { c.Args = cobra.ExactArgs(n) }
}

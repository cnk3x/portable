package flags

import (
	"cmp"
)

// Option option
type Option func(*Command)

// Options options
func Options(options ...Option) Option {
	return func(c *Command) {
		for _, option := range options {
			option(c)
		}
	}
}

// Use is the one-line usage message.
// Recommended syntax is as follows:
//
//	[ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
//	... indicates that you can specify multiple values for the previous argument.
//	|   indicates mutually exclusive information. You can use the argument to the left of the separator or the
//	    argument to the right of the separator. You cannot use both arguments in a single use of the command.
//	{ } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
//	    optional, they are enclosed in brackets ([ ]).
//
// Example: add [-F file | -D dir]... [-f format] profile
//
// Aliases is an array of aliases that can be used instead of the first word in Use.
func Use(use string, aliases ...string) Option {
	return func(c *Command) { c.Use, c.Aliases = use, aliases }
}

// Aliases is an array of aliases that can be used instead of the first word in Use.
func Aliases(aliases ...string) Option {
	return func(c *Command) { c.Aliases = aliases }
}

// Description
//
//	Short is the short description shown in the 'help' output.
//	Long is the long message shown in the 'help <this-command>' output.
func Description(short string, long ...string) Option {
	return func(c *Command) {
		c.Short = short
		if len(long) > 0 {
			c.Long = cmp.Or(long...)
		}
	}
}

// SuggestFor is an array of command names for which this command will be suggested -
// similar to aliases but only suggests.
func SuggestFor(suggestFor ...string) Option {
	return func(c *Command) { c.SuggestFor = suggestFor }
}

// The group id under which this subcommand is grouped in the 'help' output of its parent.
func GroupID(groupID string) Option {
	return func(c *Command) { c.GroupID = groupID }
}

// Deprecated defines, if this command is deprecated and should print this string when used.
func Deprecated(deprecated string) Option {
	return func(c *Command) { c.Deprecated = deprecated }
}

// Example is examples of how to use the command.
func Example(example string) Option {
	return func(c *Command) { c.Example = example }
}

// Flags returns the complete FlagSet that applies
// to this command (local and persistent declared here and by all parents).
func Flags(set func(*FlagSet)) Option {
	return func(c *Command) {
		set(c.Flags())
	}
}

// PersistentFlags returns the persistent FlagSet specifically set in the current command.
func PersistentFlags(set func(*FlagSet)) Option {
	return func(c *Command) {
		set(c.PersistentFlags())
	}
}

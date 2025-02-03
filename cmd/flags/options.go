package flags

import (
	"cmp"
	"net"
	"time"
)

type (
	Option     func(*Command) // Option option
	FlagOption func(*FlagSet) // FlagOption flag option
)

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

func Hidden(hide bool) Option { return func(c *Command) { c.Hidden = hide } }

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
func Flags(sets ...FlagOption) Option {
	return func(c *Command) {
		for _, set := range sets {
			set(c.Flags())
		}
	}
}

// PersistentFlags returns the persistent FlagSet specifically set in the current command.
func PersistentFlags(sets ...FlagOption) Option {
	return func(c *Command) {
		for _, set := range sets {
			set(c.PersistentFlags())
		}
	}
}

// Var is a helper function to add a variable to the flag set.
func Var[T VarT](name, short string, val T, usage string) FlagOption {
	return Val(&val, name, short, usage)
}

// Val is a helper function to add a variable to the flag set.
func Val[T VarT](val *T, name, short string, usage string) FlagOption {
	return func(fs *FlagSet) {
		switch x := any(val).(type) {
		case *bool:
			fs.BoolVarP(x, name, short, *x, usage)
		case *string:
			fs.StringVarP(x, name, short, *x, usage)
		case *int:
			fs.IntVarP(x, name, short, *x, usage)
		case *int8:
			fs.Int8VarP(x, name, short, *x, usage)
		case *int16:
			fs.Int16VarP(x, name, short, *x, usage)
		case *int32:
			fs.Int32VarP(x, name, short, *x, usage)
		case *int64:
			fs.Int64VarP(x, name, short, *x, usage)
		case *uint:
			fs.UintVarP(x, name, short, *x, usage)
		case *uint8:
			fs.Uint8VarP(x, name, short, *x, usage)
		case *uint16:
			fs.Uint16VarP(x, name, short, *x, usage)
		case *uint32:
			fs.Uint32VarP(x, name, short, *x, usage)
		case *uint64:
			fs.Uint64VarP(x, name, short, *x, usage)
		case *float32:
			fs.Float32VarP(x, name, short, *x, usage)
		case *float64:
			fs.Float64VarP(x, name, short, *x, usage)
		case *net.IP:
			fs.IPVarP(x, name, short, *x, usage)
		case *time.Duration:
			fs.DurationVarP(x, name, short, *x, usage)
		case *[]bool:
			fs.BoolSliceVarP(x, name, short, *x, usage)
		case *[]string:
			fs.StringSliceVarP(x, name, short, *x, usage)
		case *[]int:
			fs.IntSliceVarP(x, name, short, *x, usage)
		case *[]int32:
			fs.Int32SliceVarP(x, name, short, *x, usage)
		case *[]int64:
			fs.Int64SliceVarP(x, name, short, *x, usage)
		case *[]uint:
			fs.UintSliceVarP(x, name, short, *x, usage)
		case *[]float32:
			fs.Float32SliceVarP(x, name, short, *x, usage)
		case *[]float64:
			fs.Float64SliceVarP(x, name, short, *x, usage)
		case *[]net.IP:
			fs.IPSliceVarP(x, name, short, *x, usage)
		case *[]time.Duration:
			fs.DurationSliceVarP(x, name, short, *x, usage)
		}
	}
}

// VarT is the type of the variable to be added to the flag set.
type VarT interface {
	int8 |
		int16 | uint8 | uint16 | uint32 | uint64 |

		bool | string | int | int32 | int64 | uint | float32 | float64 | time.Duration | net.IP |

		[]bool | []string | []int | []int32 | []int64 | []uint | []float32 | []float64 | []time.Duration | []net.IP
}

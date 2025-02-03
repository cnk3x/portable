package flags

// CobraRun cobra run
type CobraRun interface {
	func() | func(args []string) | func(*Command, []string)
}

func cobraRun[F CobraRun](fn F) func(*Command, []string) {
	return func(cmd *Command, args []string) {
		switch f := any(fn).(type) {
		case func():
			f()
		case func(args []string):
			f(args)
		case func(*Command, []string):
			f(cmd, args)
		}
	}
}

// Run run the command
func Run[T CobraRun](fn T) Option {
	return func(c *Command) { c.Run = cobraRun(fn) }
}

// PreRun pre run the command
func PreRun[T CobraRun](fn T) Option {
	return func(c *Command) { c.PreRun = cobraRun(fn) }
}

// PostRun post run the command
func PostRun[T CobraRun](fn T) Option {
	return func(c *Command) { c.PostRun = cobraRun(fn) }
}

// PersistentPreRun persistent pre run the command
func PersistentPreRun[T CobraRun](fn T) Option {
	return func(c *Command) { c.PersistentPreRun = cobraRun(fn) }
}

// PersistentPostRun persistent post run the command
func PersistentPostRun[T CobraRun](fn T) Option {
	return func(c *Command) { c.PersistentPostRun = cobraRun(fn) }
}

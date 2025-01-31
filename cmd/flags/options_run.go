package flags

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

func Run[T CobraRun](fn T) Option {
	return func(c *Command) { c.Run = cobraRun(fn) }
}

func PreRun[T CobraRun](fn T) Option {
	return func(c *Command) { c.PreRun = cobraRun(fn) }
}

func PostRun[T CobraRun](fn T) Option {
	return func(c *Command) { c.PostRun = cobraRun(fn) }
}

func PersistentPreRun[T CobraRun](fn T) Option {
	return func(c *Command) { c.PersistentPreRun = cobraRun(fn) }
}

func PersistentPostRun[T CobraRun](fn T) Option {
	return func(c *Command) { c.PersistentPostRun = cobraRun(fn) }
}

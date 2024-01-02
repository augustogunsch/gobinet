package context

import (
	"log"

	"github.com/augustogunsch/gobinet/internal/args"
	"github.com/augustogunsch/gobinet/internal/notifier"
)

type Notify interface {
	Notify(*log.Logger, string)
}

type Context struct {
	N    Notify
	L    *log.Logger
	Args *args.ArgSet
}

func DefaultWithArgs(args *args.ArgSet) Context {
	return Context{
		N:    notifier.Default(),
		L:    log.Default(),
		Args: args,
	}
}

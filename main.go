package main

import (
	"fmt"
	"os"

	"github.com/augustogunsch/gobinet/internal/args"
	"github.com/augustogunsch/gobinet/internal/cmds"
	"github.com/augustogunsch/gobinet/internal/context"
)

const VERSION = "v1.1.2"

func main() {
	parsedArgs, err := args.ParseArgs()

	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		args.WriteUsage(os.Stderr)
		os.Exit(1)
	}

	if parsedArgs.Help {
		args.WriteUsage(os.Stdout)
		return
	}

	if parsedArgs.Version {
		fmt.Println("Gobinet", VERSION)
		return
	}

	ctx := context.DefaultWithArgs(&parsedArgs)

	switch parsedArgs.Cmd {
	case "build":
		cmds.Build(ctx)
	case "watch":
		cmds.Build(ctx)
		cmds.Watch(ctx)
	default:
		msg := fmt.Sprintf("unrecognized command `%s`", parsedArgs.Cmd)
		os.Stderr.WriteString(msg)
		args.WriteUsage(os.Stderr)
		os.Exit(1)
	}
}

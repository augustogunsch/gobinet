package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/augustogunsch/gobinet/internal/args"
	"github.com/augustogunsch/gobinet/internal/cmds"
	"github.com/augustogunsch/gobinet/internal/logic"
)

type notifier struct{}

func (n *notifier) Notify(l *log.Logger, msg string) {
	cmd := exec.Command("notify-send", "Gobinet error", msg)
	if output, err := cmd.CombinedOutput(); err != nil {
		l.Printf("error sending notification:\n%s", output)
		return
	}
}

const VERSION = "v1.1.1"

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

	ctx := logic.Context{
		N:    &notifier{},
		L:    log.Default(),
		Args: &parsedArgs,
	}

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

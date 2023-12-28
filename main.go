package main

import (
	"fmt"
	"os"
)

const VERSION = "v1.1.0"

func main() {
	args := parseArgs()

	switch args.Cmd {
	case "build":
		build(args)
	case "watch":
		build(args)
		watch(args)
	default:
		err := fmt.Sprintf("unrecognized command `%s`", args.Cmd)
		os.Stderr.WriteString(err)
		usage(os.Stderr, 1)
	}
}

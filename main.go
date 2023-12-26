package main

import (
	"flag"
	"os"
	"path/filepath"
)

func main() {
	args := ArgSet{}

	flag.BoolVar(&args.Help, "help", false, "show this help message and exit")

	flag.Parse()

	if args.Help {
		usage(os.Stderr, 0)
	}

	posArgs := flag.Args()

	if len(posArgs) != 3 {
		usage(os.Stderr, 1)
	}

	args.Cmd = posArgs[0]
	args.Input = filepath.Clean(posArgs[1])
	args.Output = filepath.Clean(posArgs[2])

	err := RunCmd(args)

	if err != nil {
		err.panic()
	}
}

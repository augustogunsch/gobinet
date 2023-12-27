package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type IncludeDirs []string

func (dirs *IncludeDirs) String() string {
	return strings.Join(*dirs, ":") + ":"
}

func (dirs *IncludeDirs) Set(value string) error {
	*dirs = append(*dirs, value)
	return nil
}

type ArgSet struct {
	Help    bool
	Include IncludeDirs
	Reload  bool
	Notify  bool
	Cmd     string
	Input   string
	Output  string
}

func usage(stream io.Writer, exitCode int) {
	program := path.Base(os.Args[0])
	stream.Write([]byte(fmt.Sprintf(
		"usage: %s [--help] [--include DIR] [--reload] [--notify] <build|watch> INPUT OUTPUT\n",
		program,
	)))
	flag.CommandLine.SetOutput(stream)
	flag.PrintDefaults()
	os.Exit(exitCode)
}

func parseArgs() ArgSet {
	args := ArgSet{}

	flag.BoolVar(
		&args.Help,
		"help",
		false,
		"Show this help message and exit.",
	)
	flag.BoolVar(
		&args.Reload,
		"reload",
		false,
		"Reload MuPDF by sending a HUP signal when files are updated.",
	)
	flag.Var(
		&args.Include,
		"include",
		"Include this directory. May be passed multiple times.",
	)
	flag.BoolVar(
		&args.Notify,
		"notify",
		false,
		"Send a desktop notification when compilation fails.",
	)

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

	return args
}

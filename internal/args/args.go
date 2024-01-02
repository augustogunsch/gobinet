package args

import (
	"errors"
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
	Include IncludeDirs
	Reload  bool
	Notify  bool
	Cmd     string
	Input   string
	Output  string
	Version bool
	Help    bool
}

func WriteUsage(stream io.Writer) {
	program := path.Base(os.Args[0])
	stream.Write([]byte(fmt.Sprintf(
		"usage: %s [--help] [--version] [--include DIR] [--reload] [--notify] <build|watch> INPUT OUTPUT\n",
		program,
	)))
	flag.CommandLine.SetOutput(stream)
	flag.PrintDefaults()
}

func ParseArgs() (ArgSet, error) {
	args := ArgSet{}

	flag.BoolVar(
		&args.Help,
		"help",
		false,
		"Show this help message and exit.",
	)
	flag.BoolVar(
		&args.Version,
		"version",
		false,
		"Show Gobinet's version.",
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

	if args.Help || args.Version {
		return args, nil
	}

	posArgs := flag.Args()

	if len(posArgs) != 3 {
		return args, errors.New("expected three positional arguments")
	}

	args.Cmd = posArgs[0]
	args.Input = filepath.Clean(posArgs[1])
	args.Output = filepath.Clean(posArgs[2])

	return args, nil
}

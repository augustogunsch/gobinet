package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
)

type ArgSet struct {
	Help   bool
	Cmd    string
	Input  string
	Output string
}

func usage(stream io.Writer, exitCode int) {
	program := path.Base(os.Args[0])
	stream.Write([]byte(fmt.Sprintf(
		"usage: %s [--help] CMD INPUT OUTPUT\n",
		program,
	)))
	stream.Write([]byte(fmt.Sprintf(
		"  CMD\n        value can be `build`\n",
	)))
	flag.CommandLine.SetOutput(stream)
	flag.PrintDefaults()
	os.Exit(exitCode)
}

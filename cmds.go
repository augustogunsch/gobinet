package main

import (
	"fmt"
	"os"
)

type CmdError struct {
	code int
	msg  string
}

func (err CmdError) Error() string {
	return err.msg
}

func (err CmdError) panic() {
	os.Stderr.WriteString(err.msg)
	os.Stderr.WriteString("\n")
	usage(os.Stderr, err.code)
}

func CmdErrorf(code int, format string, a ...any) *CmdError {
	err := CmdError{
		code: code,
		msg:  fmt.Sprintf(format, a...),
	}
	return &err
}

func RunCmd(args ArgSet) *CmdError {
	switch args.Cmd {
	case "build":
		build(args)
	default:
		return CmdErrorf(2, "unrecognized command `%s`", args.Cmd)
	}

	return nil
}

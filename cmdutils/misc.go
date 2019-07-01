package cmdutils

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

// ExitCode wraps parent error and append exit code.
type ExitCode struct {
	error
	// exit code.
	Code int
}

var InvalidArgument = ExitCode{
	error: errors.New("invalid argument"),
	Code:  128,
}

func RunCommand(command *cobra.Command) int {
	if err := command.Execute(); err != nil {
		if e, ok := err.(*ExitCode); ok {
			return e.Code
		}
		return 1
	}
	return 0
}

func ShowError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
}

func HandlePanic(fn func() error) (err error) {
	defer func() {
		if obj := recover(); obj != nil {
			err = obj.(error)
		}
		err = errors.WithStack(err)
	}()
	err = fn()
	return
}

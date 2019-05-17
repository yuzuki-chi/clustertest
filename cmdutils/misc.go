package cmdutils

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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

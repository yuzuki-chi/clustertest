package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	"github.com/yuuki0xff/clustertest/rpc"
	"os"
)

func taskOutputFn(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		err := errors.New("no TaskID specified")
		ShowError(err)
		return nil
	}

	c, err := rpc.NewClient()
	if err != nil {
		ShowError(err)
		return nil
	}

	taskIDs := args
	var render resultRender
	if len(taskIDs) > 1 {
		render = &multipleResultRender{}
	} else {
		render = &singleResultRender{}
	}

	for _, sid := range taskIDs {
		id := &StringTaskID{sid}
		d, err := c.Inspect(id)
		if err != nil {
			ShowError(err)
			return nil
		}

		render.Render(os.Stdout, d)
	}
	return nil
}

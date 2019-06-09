package main

import (
	"context"
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	"github.com/yuuki0xff/clustertest/rpc"
)

func taskWaitFn(cmd *cobra.Command, args []string) error {
	taskIDs := args

	c, err := rpc.NewClient()
	if err != nil {
		ShowError(err)
		return nil
	}

	ctx := context.Background()
	for _, sid := range taskIDs {
		id := &StringTaskID{sid}
		err := c.Wait(id, ctx)
		if err != nil {
			ShowError(err)
			return nil
		}
	}
	return nil
}

type StringTaskID struct {
	id string
}

func (s *StringTaskID) String() string {
	return s.id
}

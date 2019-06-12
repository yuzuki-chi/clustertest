package main

import (
	"fmt"
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	"github.com/yuuki0xff/clustertest/rpc"
)

func taskStartFn(cmd *cobra.Command, args []string) error {
	c, err := rpc.NewClient()
	if err != nil {
		ShowError(err)
		return nil
	}

	files, err := findConfigs(args)
	if err != nil {
		ShowError(err)
		return nil
	}

	for _, file := range files {
		task, err := newTaskFromFile(file)
		if err != nil {
			ShowError(err)
			return nil
		}

		id, err := c.Create(task)
		if err != nil {
			ShowError(err)
			return nil
		}
		fmt.Println(id)
	}
	return nil
}

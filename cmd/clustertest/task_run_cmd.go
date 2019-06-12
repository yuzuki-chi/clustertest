package main

import (
	"context"
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/clustertest/rpc"
	"os"
)

func taskRunFn(cmd *cobra.Command, args []string) error {
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

	var ids []models.TaskID
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
		ids = append(ids, id)
	}

	for _, id := range ids {
		err := c.Wait(id, context.Background())
		if err != nil {
			ShowError(err)
			return nil
		}
	}

	var render resultRender
	if len(ids) > 1 {
		render = &multipleResultRender{}
	} else {
		render = &singleResultRender{}
	}

	for _, id := range ids {
		d, err := c.Inspect(id)
		if err != nil {
			ShowError(err)
			return nil
		}

		render.Render(os.Stdout, d)
	}
	return nil
}

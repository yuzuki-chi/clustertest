package main

import (
	"github.com/rgeoghegan/tabulate"
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/clustertest/rpc"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

func taskListFn(cmd *cobra.Command, args []string) error {
	c, err := rpc.NewClient()
	if err != nil {
		ShowError(err)
		return nil
	}

	tasks, err := c.List()
	if err != nil {
		ShowError(err)
		return nil
	}

	// Sort by ID.
	sort.SliceStable(tasks, func(i, j int) bool {
		left := tasks[i].TaskID().String()
		right := tasks[j].TaskID().String()

		a, err1 := strconv.Atoi(left)
		b, err2 := strconv.Atoi(right)
		if err1 != nil || err2 != nil {
			// Failed convert to number.
			return strings.Compare(left, right) < 0
		}
		return a < b
	})

	taskListRender{}.Render(os.Stdout, tasks)
	return nil
}

type taskListRender struct{}

func (taskListRender) Render(w io.Writer, tasks []models.TaskDetail) {
	var rows []*taskListRow
	for _, t := range tasks {
		row := &taskListRow{
			ID:     t.TaskID().String(),
			Status: t.State(),
		}
		rows = append(rows, row)
	}

	layout := &tabulate.Layout{
		Format: tabulate.SimpleFormat,
	}
	table, err := tabulate.Tabulate(rows, layout)
	if err != nil {
		panic(err)
	}
	io.WriteString(w, table)
}

type taskListRow struct {
	ID     string
	Status string
}

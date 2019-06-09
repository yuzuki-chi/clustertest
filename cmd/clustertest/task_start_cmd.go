package main

import (
	"fmt"
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/clustertest/rpc"
	"io/ioutil"
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

func newTaskFromFile(name string) (models.Task, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	return &FileTask{name, data}, nil
}

type FileTask struct {
	name string
	data []byte
}

func (t *FileTask) String() string {
	return t.name
}
func (t *FileTask) SpecData() []byte {
	return t.data
}

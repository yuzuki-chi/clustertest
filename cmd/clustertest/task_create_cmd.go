package main

import (
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	"github.com/yuuki0xff/clustertest/config"
)

func taskCreateFn(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		// Load a config file from "./".
		args = []string{"./"}
	}
	tasks, err := config.LoadFromDirsOrFiles(args)
	if err != nil {
		ShowError(err)
		return nil
	}
	// TODO: enqueue tasks
	_ = tasks
	return nil
}

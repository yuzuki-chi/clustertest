package main

import (
	"github.com/spf13/cobra"
	"github.com/yuuki0xff/clustertest/cmdutils"
)

func taskStartFn(cmd *cobra.Command, args []string) error {
	confs, err := loadConfigs(args)
	if err != nil {
		cmdutils.ShowError(err)
		return nil
	}

	for _, conf := range confs {
		_ = conf
		// TODO: enqueue conf
	}
}

package main

import (
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	"github.com/yuuki0xff/clustertest/config"
	"github.com/yuuki0xff/clustertest/executors"
	"github.com/yuuki0xff/clustertest/provisioners"
	_ "github.com/yuuki0xff/clustertest/provisioners/proxmoxve"
	_ "github.com/yuuki0xff/clustertest/scripts/localshell"
	"os"
)

func taskCreateFn(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		// Load a config file from "./".
		args = []string{"./"}
	}
	confs, err := config.LoadFromDirsOrFiles(args)
	if err != nil {
		ShowError(err)
		return nil
	}

	// TODO: enqueue tasks

	// DEBUG: create resources
	for _, conf := range confs {
		for _, s := range conf.Specs() {
			pro, err := provisioners.New(s)
			if err != nil {
				ShowError(err)
				return nil
			}

			err = pro.Create()
			if err != nil {
				ShowError(err)
				return nil
			}

			sets := pro.ScriptSets()
			r := executors.ExecuteBefore(pro, sets)
			os.Stdout.Write(r.Output())
			r = executors.ExecuteMain(pro, sets)
			os.Stdout.Write(r.Output())
			r = executors.ExecuteAfter(pro, sets)
			os.Stdout.Write(r.Output())

			err = pro.Delete()
			if err != nil {
				ShowError(err)
				return nil
			}
		}
	}

	return nil
}

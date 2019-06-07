package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	"github.com/yuuki0xff/clustertest/executors"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/clustertest/provisioners"
	"os"
)

func taskRunFn(cmd *cobra.Command, args []string) error {
	confs, err := loadConfigs(args)
	if err != nil {
		ShowError(err)
		return nil
	}

	for _, conf := range confs {
		// Create provisioners.
		var pros []models.Provisioner
		for _, s := range conf.Specs() {
			pro, err := provisioners.New(s)
			if err != nil {
				ShowError(err)
				return nil
			}
			pros = append(pros, pro)
		}

		// Create resources.
		for _, pro := range pros {
			err = pro.Create()
			if err != nil {
				ShowError(err)
				return nil
			}
		}

		// Run scripts.
		for _, pro := range pros {
			sets := pro.ScriptSets()
			r := executors.ExecuteBefore(pro, sets)
			os.Stdout.Write(r.Output())
			if r.ExitCode() != 0 {
				ShowError(errors.Errorf("failed the \"before\" task: exitcode=%d", r.ExitCode()))
				return nil
			}
			r = executors.ExecuteMain(pro, sets)
			os.Stdout.Write(r.Output())
			if r.ExitCode() != 0 {
				ShowError(errors.Errorf("failed the \"main\" task: exitcode=%d", r.ExitCode()))
				return nil
			}
			r = executors.ExecuteAfter(pro, sets)
			os.Stdout.Write(r.Output())
			if r.ExitCode() != 0 {
				ShowError(errors.Errorf("failed the \"after\" task: exitcode=%d", r.ExitCode()))
				return nil
			}
		}

		// Delete resources.
		for _, pro := range pros {
			err = pro.Delete()
			if err != nil {
				ShowError(err)
				return nil
			}
		}
	}

	return nil
}

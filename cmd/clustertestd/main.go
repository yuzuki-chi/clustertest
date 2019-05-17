package main

import (
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	"os"
)

var rootCmd = &cobra.Command{
	Use:              "clustertestd",
	Short:            "Clustertest server",
	TraverseChildren: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO
		panic("not implemented")
	},
}

func main() {
	os.Exit(RunCommand(rootCmd))
}

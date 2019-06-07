package main

import (
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	_ "github.com/yuuki0xff/clustertest/import_all"
	"os"
)

var rootCmd = &cobra.Command{
	Use:              "clustertestd",
	Short:            "Start clustertest daemon",
	TraverseChildren: true,
	RunE:             rootCmdFn,
}
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current daemon status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO
		panic("not implemented")
	},
}
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "daemon going to graceful shutdown",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO
		panic("not implemented")
	},
}
var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Stop daemon immediately",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO
		panic("not implemented")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd, stopCmd, killCmd)
}

func main() {
	os.Exit(RunCommand(rootCmd))
}

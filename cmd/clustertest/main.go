package main

import (
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	"os"
)

func notImplemented(cmd *cobra.Command, args []string) error {
	// TODO
	panic("not implemented")
}

var rootCmd = &cobra.Command{
	Use:              "clustertest",
	Short:            "An automated testing system for clustered system",
	TraverseChildren: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return InvalidArgument
	},
}
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		return InvalidArgument
	},
}
var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a task and enqueue it",
	RunE:  notImplemented,
}
var taskWaitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Wait for task to done",
	RunE:  notImplemented,
}
var taskCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel tasks",
	RunE:  notImplemented,
}
var taskOutputCmd = &cobra.Command{
	Use:   "output",
	Short: "Show output data of a task",
	RunE:  notImplemented,
}
var taskDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete finished tasks",
	RunE:  notImplemented,
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskCreateCmd, taskWaitCmd, taskCancelCmd, taskOutputCmd, taskDeleteCmd)
}

func main() {
	os.Exit(RunCommand(rootCmd))
}

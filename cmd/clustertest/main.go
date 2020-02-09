package main

import (
	"github.com/spf13/cobra"
	. "github.com/yuuki0xff/clustertest/cmdutils"
	_ "github.com/yuuki0xff/clustertest/import_all"
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
var taskRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a task and wait it",
	RunE:  taskRunFn,
}
var taskStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a task in background",
	RunE:  taskStartFn,
}
var taskWaitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Wait for task to done",
	RunE:  taskWaitFn,
}
var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE:  taskListFn,
}
var taskCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel tasks",
	RunE:  notImplemented,
}
var taskOutputCmd = &cobra.Command{
	Use:   "output",
	Short: "Show output data of a task",
	RunE:  taskOutputFn,
}
var taskDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete finished tasks",
	RunE:  taskDeleteFn,
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskRunCmd, taskStartCmd, taskWaitCmd, taskListCmd, taskCancelCmd, taskOutputCmd, taskDeleteCmd)
}

func main() {
	os.Exit(RunCommand(rootCmd))
}

package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/yuuki0xff/clustertest/rpc"
)

func rootCmdFn(cmd *cobra.Command, args []string) error {
	addr := "0.0.0.0:9571"
	fmt.Printf("Listening on %s\n", addr)
	return rpc.ServeServer(addr)
}

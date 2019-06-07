package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/yuuki0xff/clustertest/databases"
	"github.com/yuuki0xff/clustertest/rpc"
	"github.com/yuuki0xff/clustertest/worker"
	"golang.org/x/sync/errgroup"
)

func rootCmdFn(cmd *cobra.Command, args []string) error {
	db := databases.NewMemTaskDB()

	ctx := context.Background()
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		w := worker.Worker{
			Queue: db,
		}
		return w.Serve(ctx)
	})
	g.Go(func() error {
		addr := "0.0.0.0:9571"
		fmt.Printf("Listening on %s\n", addr)
		return rpc.ServeServer(addr, db)
	})
	return g.Wait()
}

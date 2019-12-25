package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/yuuki0xff/clustertest/databases"
	"github.com/yuuki0xff/clustertest/rpc"
	"github.com/yuuki0xff/clustertest/worker"
	"golang.org/x/sync/errgroup"
	"os"
)

func rootCmdFn(cmd *cobra.Command, args []string) error {
	jobs, err := cmd.Flags().GetInt32("jobs")
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: invalid args", err)
		return nil
	}
	if jobs <= 0 {
		fmt.Fprintln(os.Stderr, "ERROR: --jobs must be larger than 0")
		return nil
	}

	listen, err := cmd.Flags().GetString("listen")
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: invalid args", err)
		return nil
	}

	port, err := cmd.Flags().GetInt32("port")
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: invalid args", err)
		return nil
	}
	if port <= 0 {
		fmt.Fprintln(os.Stderr, "ERROR: --port must be larger than 0")
		return nil
	}

	db := databases.NewMemTaskDB()

	ctx := context.Background()
	g, _ := errgroup.WithContext(ctx)
	for i := int32(0); i < jobs; i++ {
		g.Go(func() error {
			w := worker.Worker{
				Queue: db,
			}
			return w.Serve(ctx)
		})
	}
	g.Go(func() error {
		addr := fmt.Sprintf("%s:%d", listen, port)
		fmt.Printf("Listening on %s\n", addr)
		return rpc.ServeServer(addr, db)
	})
	return g.Wait()
}

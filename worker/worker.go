package worker

import (
	"context"
	"github.com/pkg/errors"
	"github.com/yuuki0xff/clustertest/models"
	"time"
)

type Worker struct {
	Queue models.TaskQueue
}

func (w *Worker) Serve(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("context canceled")

		case <-ticker.C:
			err := w.Queue.Consume(w.runTask)
			if err != nil {
				if err == models.QueueEmpty {
					continue
				}
				return err
			}
		}
	}
}
func (w *Worker) runTask(id models.TaskID, task models.Task) (models.TaskResult, error) {
	// todo
	panic("not impl")
}

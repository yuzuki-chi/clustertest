package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/yuuki0xff/clustertest/config"
	"github.com/yuuki0xff/clustertest/executors"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/clustertest/provisioners"
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
	result := &Result{}
	fmt.Println("running", id, result)
	defer func() {
		js, _ := json.Marshal(result)
		fmt.Println("finished", id, string(js))
	}()

	data := task.SpecData()
	conf, err := config.LoadFromBytes(data)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("failed to load spec: %s", err)
		return result, nil
	}

	// Create provisioners.
	var pros []models.Provisioner
	for _, s := range conf.Specs() {
		pro, err := provisioners.New(s)
		if err != nil {
			result.ErrorMsg = err.Error()
			return result, nil
		}
		pros = append(pros, pro)
	}

	// Create resources.
	for _, pro := range pros {
		err = pro.Create()
		if err != nil {
			result.ErrorMsg = err.Error()
			return result, nil
		}
	}

	// Run scripts.
	before := executors.MergedResult{}
	main := executors.MergedResult{}
	after := executors.MergedResult{}

	for _, pro := range pros {
		sets := pro.ScriptSets()
		r := executors.ExecuteBefore(pro, sets)
		before.Append(r)
		if r.ExitCode() != 0 {
			result.ErrorMsg = fmt.Sprintf("failed the \"before\" task: exitcode=%d", r.ExitCode())
			result.Before = NewScriptResult(&before)
			return result, nil
		}
	}
	result.Before = NewScriptResult(&before)

	for _, pro := range pros {
		sets := pro.ScriptSets()
		r := executors.ExecuteMain(pro, sets)
		main.Append(r)
		if r.ExitCode() != 0 {
			result.ErrorMsg = fmt.Sprintf("failed the \"main\" task: exitcode=%d", r.ExitCode())
			result.Main = NewScriptResult(&main)
			return result, nil
		}
	}
	result.Main = NewScriptResult(&main)

	for _, pro := range pros {
		sets := pro.ScriptSets()
		r := executors.ExecuteAfter(pro, sets)
		after.Append(r)
		if r.ExitCode() != 0 {
			result.ErrorMsg = fmt.Sprintf("failed the \"after\" task: exitcode=%d", r.ExitCode())
			result.After = NewScriptResult(&after)
			return result, nil
		}
	}
	result.After = NewScriptResult(&after)

	// Delete resources.
	for _, pro := range pros {
		err = pro.Delete()
		if err != nil {
			result.ErrorMsg = err.Error()
			return result, nil
		}
	}

	return result, nil
}

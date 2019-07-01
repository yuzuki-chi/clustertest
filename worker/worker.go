package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/republicprotocol/co-go"
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
		pro, err := provisioners.New(conf.Name, s)
		if err != nil {
			result.ErrorMsg = err.Error()
			return result, nil
		}
		pros = append(pros, pro)
	}

	// Reserve resources.
	ec := make(chan error, len(pros))
	co.ParForAll(pros, func(i int) {
		err := pros[i].Reserve()
		if err != nil {
			ec <- err
		}
	})
	close(ec)
	if err = <-ec; err != nil {
		result.ErrorMsg = err.Error()
		return result, nil
	}

	// Create resources.
	ec = make(chan error, len(pros))
	co.ParForAll(pros, func(i int) {
		err := pros[i].Create()
		if err != nil {
			ec <- err
		}
	})
	close(ec)
	if err = <-ec; err != nil {
		result.ErrorMsg = err.Error()
		return result, nil
	}

	// Run the "before" script.
	before := executors.MergedResult{}
	rc := make(chan models.ScriptResult, len(pros))
	co.ParForAll(pros, func(i int) {
		pro := pros[i]
		sets := pro.ScriptSets()
		rc <- executors.ExecuteBefore(pro, sets)
	})
	close(rc)
	for r := range rc {
		before.Append(r)
		if r.ExitCode() != 0 {
			result.ErrorMsg = fmt.Sprintf("failed the \"before\" task: exitcode=%d", r.ExitCode())
			result.Before = NewScriptResult(&before)
			return result, nil
		}
	}
	result.Before = NewScriptResult(&before)

	// Run the "main" script.
	main := executors.MergedResult{}
	rc = make(chan models.ScriptResult, len(pros))
	co.ParForAll(pros, func(i int) {
		pro := pros[i]
		sets := pro.ScriptSets()
		rc <- executors.ExecuteMain(pro, sets)
	})
	close(rc)
	for r := range rc {
		main.Append(r)
		if r.ExitCode() != 0 {
			result.ErrorMsg = fmt.Sprintf("failed the \"main\" task: exitcode=%d", r.ExitCode())
			result.Main = NewScriptResult(&main)
			return result, nil
		}
	}
	result.Main = NewScriptResult(&main)

	// Run the "after" script.
	after := executors.MergedResult{}
	rc = make(chan models.ScriptResult, len(pros))
	co.ParForAll(pros, func(i int) {
		pro := pros[i]
		sets := pro.ScriptSets()
		rc <- executors.ExecuteAfter(pro, sets)
	})
	close(rc)
	for r := range rc {
		after.Append(r)
		if r.ExitCode() != 0 {
			result.ErrorMsg = fmt.Sprintf("failed the \"after\" task: exitcode=%d", r.ExitCode())
			result.After = NewScriptResult(&after)
			return result, nil
		}
	}
	result.After = NewScriptResult(&after)

	// Delete resources.
	ec = make(chan error, len(pros))
	co.ParForAll(pros, func(i int) {
		pro := pros[i]
		err := pro.Delete()
		if err != nil {
			ec <- err
		}
	})
	close(ec)
	if err = <-ec; err != nil {
		result.ErrorMsg = err.Error()
		return result, nil
	}

	return result, nil
}

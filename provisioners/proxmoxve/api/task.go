package api

import (
	"context"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type TaskID string
type TaskStatus struct {
	Status     string
	ExitStatus string `json:"exitstatus"`
}
type Task struct {
	TaskID TaskID
	NodeID NodeID
	Client *PveClient
}

func (s *TaskStatus) IsStopped() bool {
	return s.Status == "stopped"
}
func (s *TaskStatus) IsOK() bool {
	return s.ExitStatus == "OK"
}

func (t *Task) Wait(ctx context.Context) error {
	var status TaskStatus
	timer := time.NewTimer(0)

	// Wait for task to complete.
WaitLoop:
	for {
		timer.Reset(time.Second)

		select {
		case <-timer.C:
			var err error
			status, err = t.Client.taskStatus(t)
			if err != nil {
				return err
			}
			if status.IsStopped() {
				break WaitLoop
			}
		case <-ctx.Done():
			return errors.Errorf("task timeout")
		}
	}

	// Check the result.
	if status.IsOK() {
		return nil
	}

	lines, err := t.LogAll()
	if err != nil {
		return err
	}
	return errors.Errorf("task failed: %s", strings.Join(lines, "\n"))
}
func (t *Task) LogAll() ([]string, error) {
	limit := 1000
	var lines []string
	for offset := 0; ; offset += limit {
		ls, err := t.Client.taskLog(t, offset, limit)
		if err != nil {
			return nil, err
		}
		for _, l := range ls {
			lines = append(lines, l)
		}
		if len(ls) < limit {
			// Reached to end.
			break
		}
	}
	return lines, nil
}

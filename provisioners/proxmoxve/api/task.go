package api

import (
	"context"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
	"strings"
	"time"
)

var taskSem = semaphore.NewWeighted(4)

type TaskID string
type TaskStatus struct {
	Status     string
	ExitStatus string `json:"exitstatus"`
}
type Task struct {
	// doneFn represents that fn() is executed.
	doneFn chan struct{}
	// done represents that the Task is finished.
	done  chan struct{}
	error error

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

func NewTask(fn func(task *Task) error) *Task {
	t := &Task{
		doneFn: make(chan struct{}),
		done:   make(chan struct{}),
	}

	go func() {
		taskSem.Acquire(context.Background(), 1)
		defer taskSem.Release(1)
		defer close(t.done)

		t.error = func() error {
			defer close(t.doneFn)
			return fn(t)
		}()
		if t.error != nil {
			return
		}

		t.error = t.wait(context.Background())
	}()
	return t
}
func (t *Task) WaitFn(ctx context.Context) error {
	select {
	case <-t.doneFn:
		return t.error
	case <-ctx.Done():
		return errors.Errorf("task timeout")
	}
}
func (t *Task) Wait(ctx context.Context) error {
	select {
	case <-t.done:
		return t.error
	case <-ctx.Done():
		return errors.Errorf("task timeout: NodeID=%s TaskID=%s", t.NodeID, t.TaskID)
	}
}
func (t *Task) wait(ctx context.Context) error {
	var status TaskStatus
	ticker := time.NewTicker(time.Second)

	// Wait for task to complete.
WaitLoop:
	for {
		select {
		case <-ticker.C:
			var err error
			status, err = t.Client.taskStatus(t)
			if err != nil {
				return err
			}
			if status.IsStopped() {
				break WaitLoop
			}
		case <-ctx.Done():
			return errors.Errorf("task timeout: NodeID=%s TaskID=%s", t.NodeID, t.TaskID)
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

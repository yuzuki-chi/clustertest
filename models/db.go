package models

import (
	"context"
	"errors"
)

type TaskDB interface {
	Create(task Task) (TaskID, error)
	Inspect(id TaskID) (TaskDetail, error)
	Wait(id TaskID, ctx context.Context) error
	Cancel(id TaskID) error
	Delete(id TaskID) error
}

type TaskQueue interface {
	// Consume a task.
	// If queue is empty, it will return QueueEmpty.
	Consume(fn TaskConsumer) error
}
type TaskConsumer func(id TaskID, task Task) (TaskResult, error)

var QueueEmpty = errors.New("queue empty")

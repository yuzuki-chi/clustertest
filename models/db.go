package models

import "context"

type TaskDB interface {
	Create(task Task) (TaskID, error)
	Inspect(id TaskID) (TaskDetail, error)
	Wait(id TaskID, ctx context.Context) error
	Cancel(id TaskID) error
	Delete(id TaskID) error
}

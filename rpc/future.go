package rpc

import (
	"fmt"
	"github.com/yuuki0xff/clustertest/models"
)

type Future struct {
	id     models.TaskID
	ready  bool
	result *Result
}

func NewFuture(d models.TaskDetail) *Future {
	return &Future{
		id:     d.TaskID(),
		ready:  d.State() == "finished",
		result: NewResult(d.TaskID(), d.Result()),
	}
}
func (f *Future) String() string {
	return fmt.Sprintf("<Future %s>", f.id.String())
}
func (f *Future) TaskID() models.TaskID {
	return f.id
}
func (f *Future) State() string {
	if f.ready {
		return "finished"
	}
	return "running"
}
func (f *Future) Result() models.TaskResult {
	if f.result != nil {
		return f.result
	}
	return nil
}

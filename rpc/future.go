package rpc

import (
	"fmt"
	"github.com/yuuki0xff/clustertest/models"
)

type Detail struct {
	id     models.TaskID
	ready  bool
	result *Result
}

func NewDetail(d models.TaskDetail) *Detail {
	return &Detail{
		id:     d.TaskID(),
		ready:  d.State() == "finished",
		result: NewResult(d.TaskID(), d.Result()),
	}
}
func (f *Detail) String() string {
	return fmt.Sprintf("<Detail %s>", f.id.String())
}
func (f *Detail) TaskID() models.TaskID {
	return f.id
}
func (f *Detail) State() string {
	if f.ready {
		return "finished"
	}
	return "running"
}
func (f *Detail) Result() models.TaskResult {
	if f.result != nil {
		return f.result
	}
	return nil
}

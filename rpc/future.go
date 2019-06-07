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
	return f.result
}

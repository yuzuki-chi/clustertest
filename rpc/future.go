package rpc

import (
	"fmt"
	"github.com/yuuki0xff/clustertest/models"
)

type Detail struct {
	ID        *TaskID
	StatusStr string
	ResultObj *Result
}

func NewDetail(d models.TaskDetail) *Detail {
	return &Detail{
		ID:        NewTaskID(d.TaskID()),
		StatusStr: d.State(),
		ResultObj: NewResult(d.TaskID(), d.Result()),
	}
}
func (f *Detail) String() string {
	return fmt.Sprintf("<Detail %s>", f.ID.String())
}
func (f *Detail) TaskID() models.TaskID {
	return f.ID
}
func (f *Detail) State() string {
	return f.StatusStr
}
func (f *Detail) Result() models.TaskResult {
	if f.ResultObj != nil {
		return f.ResultObj
	}
	return nil
}

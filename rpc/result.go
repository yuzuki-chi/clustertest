package rpc

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/yuuki0xff/clustertest/models"
	"time"
)

type TaskID struct {
	id string
}
type Result struct {
	ID     TaskID
	ErrMsg string
	Before *ScriptResult
	Main   *ScriptResult
	After  *ScriptResult
}
type ScriptResult struct {
	Start    time.Time
	End      time.Time
	Hostname string
	Out      []byte
	Exit     int
}

func (t *TaskID) String() string {
	return t.id
}

func NewResult(id models.TaskID, tr models.TaskResult) *Result {
	r := &Result{
		ID:     TaskID{id.String()},
		Before: NewScriptResult(tr.BeforeResult()),
		Main:   NewScriptResult(tr.ScriptResult()),
		After:  NewScriptResult(tr.AfterResult()),
	}
	if err := tr.Error(); err != nil {
		r.ErrMsg = err.Error()
	}
	return r
}
func (r *Result) String() string {
	return fmt.Sprintf("<Result %s>", r.ID.String())
}
func (r *Result) Error() error {
	if r.ErrMsg != "" {
		return errors.New(r.ErrMsg)
	}
	return nil
}
func (r *Result) BeforeResult() models.ScriptResult {
	if r.Before != nil {
		return r.Before
	}
	return nil
}
func (r *Result) ScriptResult() models.ScriptResult {
	if r.Main != nil {
		return r.Main
	}
	return nil
}
func (r *Result) AfterResult() models.ScriptResult {
	if r.After != nil {
		return r.After
	}
	return nil
}

func NewScriptResult(r models.ScriptResult) *ScriptResult {
	if r == nil {
		return nil
	}
	return &ScriptResult{
		Start:    r.StartTime(),
		End:      r.EndTime(),
		Hostname: r.Host(),
		Out:      r.Output(),
		Exit:     r.ExitCode(),
	}
}
func (r *ScriptResult) String() string       { return "<ScriptResult>" }
func (r *ScriptResult) StartTime() time.Time { return r.Start }
func (r *ScriptResult) EndTime() time.Time   { return r.End }
func (r *ScriptResult) Host() string         { return r.Hostname }
func (r *ScriptResult) Output() []byte       { return r.Out }
func (r *ScriptResult) ExitCode() int        { return r.Exit }

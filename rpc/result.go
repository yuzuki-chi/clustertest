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
	ExitCode int
	Before   *ScriptResult
	Main     *ScriptResult
	After    *ScriptResult

	id TaskID
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

func (r *Result) String() string {
	return fmt.Sprintf("<Result %s>", r.id.String())
}
func (r *Result) Error() error {
	if r.ExitCode != 0 {
		return errors.Errorf("command failed with exit code %d", r.ExitCode)
	}
	return nil
}
func (r *Result) BeforeResult() models.ScriptResult { return r.Before }
func (r *Result) ScriptResult() models.ScriptResult { return r.Main }
func (r *Result) AfterResult() models.ScriptResult  { return r.After }

func (r *ScriptResult) String() string       { return "<ScriptResult>" }
func (r *ScriptResult) StartTime() time.Time { return r.Start }
func (r *ScriptResult) EndTime() time.Time   { return r.End }
func (r *ScriptResult) Host() string         { return r.Hostname }
func (r *ScriptResult) Output() []byte       { return r.Out }
func (r *ScriptResult) ExitCode() int        { return r.Exit }

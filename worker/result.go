package worker

import (
	"errors"
	"github.com/yuuki0xff/clustertest/models"
	"time"
)

type Result struct {
	ErrorMsg string
	Before   *ScriptResult
	Main     *ScriptResult
	After    *ScriptResult
}
type ScriptResult struct {
	Start    time.Time
	End      time.Time
	Hostname string
	Out      []byte
	Exit     int
}

func (r *Result) String() string {
	return "<Result>"
}
func (r *Result) Error() error {
	if r.ErrorMsg == "" {
		return nil
	}
	return errors.New(r.ErrorMsg)
}
func (r *Result) BeforeResult() models.ScriptResult {
	if r.Before == nil {
		return nil
	}
	return r.Before
}
func (r *Result) ScriptResult() models.ScriptResult {
	if r.Main == nil {
		return nil
	}
	return r.Main
}
func (r *Result) AfterResult() models.ScriptResult {
	if r.After == nil {
		return nil
	}
	return r.After
}

func NewScriptResult(result models.ScriptResult) *ScriptResult {
	return &ScriptResult{
		Start:    result.StartTime(),
		End:      result.EndTime(),
		Hostname: result.Host(),
		Out:      result.Output(),
		Exit:     result.ExitCode(),
	}
}
func (sr *ScriptResult) String() string {
	return "<ScriptResult>"
}
func (sr *ScriptResult) StartTime() time.Time { return sr.Start }
func (sr *ScriptResult) EndTime() time.Time   { return sr.End }
func (sr *ScriptResult) Host() string         { return sr.Hostname }
func (sr *ScriptResult) Output() []byte       { return sr.Out }
func (sr *ScriptResult) ExitCode() int        { return sr.Exit }

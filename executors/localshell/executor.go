package localshell

import (
	"bytes"
	"fmt"
	"github.com/yuuki0xff/clustertest/executors"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/clustertest/scripts/localshell"
	"os/exec"
	"time"
)

const supportedType = models.ScriptType("local-shell")

type Executor struct{}

// Result represents an result.
type Result struct {
	Command string
	Start   time.Time
	End     time.Time
	Out     []byte
	Code    int
}

func (e *Executor) String() string {
	return "<LocalShellExecutor>"
}
func (e *Executor) Type() models.ScriptType {
	return supportedType
}
func (e *Executor) Execute(script models.Script) models.ScriptResult {
	if e.Type() != script.Type() {
		panic("not supported type")
	}
	s := script.(*localshell.Script)
	return executeMany(s.Commands)
}
func executeMany(cmds []string) models.ScriptResult {
	mr := &executors.MergedResult{}
	for _, cmd := range cmds {
		result := execute(cmd)
		mr.Append(result)
		if result.ExitCode() != 0 {
			// Failed.  Stop jobs immediately.
			return mr
		}
	}
	return mr
}
func execute(cmd string) *Result {
	c := exec.Command("/bin/sh", "-c", cmd)
	r := &Result{
		Command: cmd,
		Start:   time.Now(),
	}
	out, err := c.CombinedOutput()
	r.End = time.Now()
	if _, ok := err.(*exec.ExitError); err == nil || ok {
		r.Out = out
		r.Code = c.ProcessState.ExitCode()
		return r
	}
	// Unexpected error occurred.
	r.Out = []byte(fmt.Sprintf("ERROR: %s", err.Error()))
	r.Code = 1
	return r
}

func (r *Result) String() string {
	return fmt.Sprintf("<LocalSehllResult %s>", r.Command)
}
func (r *Result) StartTime() time.Time {
	return r.Start
}
func (r *Result) Host() string {
	panic("not implemented")
}
func (r *Result) EndTime() time.Time {
	return r.End
}
func (r *Result) Output() []byte {
	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, "$ %s\n", r.Command)
	buf.Write(r.Out)
	if len(r.Out) > 0 && !bytes.HasSuffix(r.Out, []byte("\n")) {
		buf.WriteString("\n")
	}
	return buf.Bytes()
}
func (r *Result) ExitCode() int {
	return r.Code
}

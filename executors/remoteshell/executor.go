package remoteshell

import (
	"bytes"
	"fmt"
	"github.com/yuuki0xff/clustertest/executors"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/clustertest/scripts/remoteshell"
	"os/exec"
	"time"
)

const supportedType = models.ScriptType("remote-shell")

type Executor struct {
	User string
	Host string
}
type Result struct {
	E       *Executor
	Command string
	Start   time.Time
	End     time.Time
	Out     []byte
	Code    int
}

func (e *Executor) String() string {
	return "<RemoteShellExecutor>"
}
func (e *Executor) Type() models.ScriptType {
	return supportedType
}
func (e *Executor) Execute(script models.Script) models.ScriptResult {
	if e.Type() != script.Type() {
		err := fmt.Errorf("not supported type: %s does not support %s", e.Type(), script.Type())
		panic(err)
	}
	s := script.(*remoteshell.Script)
	return e.executeMany(s.Commands)
}
func (e *Executor) executeMany(cmds []string) models.ScriptResult {
	mr := &executors.MergedResult{}
	for _, cmd := range cmds {
		result := e.executeOne(cmd)
		mr.Append(result)
		if result.ExitCode() != 0 {
			// Failed.  Stop jobs immediately.
			return mr
		}
	}
	return mr
}
func (e *Executor) executeOne(cmd string) *Result {
	sshCmd := e.sshCommand(cmd)
	c := exec.Command(sshCmd[0], sshCmd[1:]...)
	r := &Result{
		E:       e,
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
func (e *Executor) sshCommand(remoteCmd string) []string {
	dest := e.sshDestinationHost()

	sshCmd := []string{"ssh"}
	// Add options to disable known_hosts.
	sshCmd = append(sshCmd, "-o", "StrictHostKeyChecking=no")
	sshCmd = append(sshCmd, "-o", "UserKnownHostsFile=/dev/null")
	sshCmd = append(sshCmd, "--", dest, remoteCmd)
	return sshCmd
}
func (e *Executor) sshDestinationHost() string {
	u := "root"
	if e.User != "" {
		u = e.User
	}
	return fmt.Sprintf("%s@%s", u, e.Host)
}
func (r *Result) String() string {
	return fmt.Sprintf("<RemoteSehllResult %s>", r.Command)
}
func (r *Result) StartTime() time.Time {
	return r.Start
}
func (r *Result) Host() string {
	return r.E.sshDestinationHost()
}
func (r *Result) EndTime() time.Time {
	return r.End
}
func (r *Result) Output() []byte {
	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, "%s$ %s\n", r.E.sshDestinationHost(), r.Command)
	buf.Write(r.Out)
	if len(r.Out) > 0 && !bytes.HasSuffix(r.Out, []byte("\n")) {
		buf.WriteString("\n")
	}
	return buf.Bytes()
}
func (r *Result) ExitCode() int {
	return r.Code
}

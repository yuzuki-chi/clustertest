package localshell

import (
	"github.com/stretchr/testify/assert"
	"github.com/yuuki0xff/clustertest/scripts/localshell"
	"testing"
)

func TestExecutor_Execute(t *testing.T) {
	t.Run("should_fail_when_passed_the_unsupported_script", func(t *testing.T) {
		assert.Panics(t, func() {
			e := Executor{}
			e.Execute(nil)
		})
	})

	t.Run("should_success_with_a_command", func(t *testing.T) {
		e := Executor{}
		s := &localshell.Script{
			Commands: []string{
				"echo foo",
			},
		}
		r := e.Execute(s)
		if !assert.NotNil(t, r) {
			return
		}
		assert.Equal(t, 0, r.ExitCode())
		assert.Equal(t, []byte(`localhost$ echo foo
foo
`), r.Output())
	})

	t.Run("should_success_with_multiple_commands", func(t *testing.T) {
		e := Executor{}
		s := &localshell.Script{
			Commands: []string{
				"echo foo",
				"echo bar",
			},
		}
		r := e.Execute(s)
		if !assert.NotNil(t, r) {
			return
		}
		assert.Equal(t, 0, r.ExitCode())
		assert.Equal(t, []byte(`localhost$ echo foo
foo
localhost$ echo bar
bar
`), r.Output())
	})

	t.Run("should_stop_when_command_failed", func(t *testing.T) {
		e := Executor{}
		s := &localshell.Script{
			Commands: []string{
				"echo foo",
				"false",
				"echo bar",
			},
		}
		r := e.Execute(s)
		if !assert.NotNil(t, r) {
			return
		}
		assert.Equal(t, 1, r.ExitCode())
		assert.Equal(t, []byte(`localhost$ echo foo
foo
localhost$ false
`), r.Output())
	})
}

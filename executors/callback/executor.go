package callback

import "github.com/yuuki0xff/clustertest/models"

const supportedType = models.ScriptType("callback")

type Callback func(script models.Script) models.ScriptResult

type Executor struct {
	Fn Callback
}

func (e *Executor) String() string {
	return "<CallbackExecutor>"
}
func (e *Executor) Type() models.ScriptType {
	return supportedType
}
func (e *Executor) Execute(script models.Script) models.ScriptResult {
	return e.Fn(script)
}

package models

import (
	"fmt"
)

type Task interface {
	fmt.Stringer
	SpecData() []byte
}
type TaskID interface {
	fmt.Stringer
}
type TaskDetail interface {
	fmt.Stringer
	TaskID() TaskID
	State() string
	Result() TaskResult
}
type TaskResult interface {
	fmt.Stringer
	Error() error
	BeforeResult() ScriptResult
	ScriptResult() ScriptResult
	AfterResult() ScriptResult
}

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
	TaskID() string
}
type TaskDetail interface {
	TaskID
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

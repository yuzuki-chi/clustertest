package models

import (
	"encoding/json"
	"fmt"
)

type Task interface {
	fmt.Stringer
	json.Marshaler
	json.Unmarshaler
	// Spec returns an infrastructure specification.
	Spec() Spec
	// Before returns an script.  It will execute before execute main script.
	// This script should prepare environment for main script (e.g. install software, create config files, build a cluster).
	Before() Script
	// Script returns an main script.
	Script() Script
	// After returns an script.  It will execute after executed main script.
	// This script should execute post processing of main script (e.g. verify the result, collect statistics data, shutdown a cluster).
	After() Script
}
type TaskID interface {
	fmt.Stringer
	TaskID() string
}
type TaskDetail interface {
	TaskID
	State()
	Result() TaskResult
}
type TaskResult interface {
	fmt.Stringer
	Error() error
	BeforeResult() ScriptResult
	ScriptResult() ScriptResult
	AfterResult() ScriptResult
}

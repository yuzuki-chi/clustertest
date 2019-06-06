package models

import (
	"fmt"
	"math"
	"time"
)

const NO_EXIT_CODE = int(math.MaxInt32)

type ScriptType string

type ScriptSet struct {
	// (Optional) It will execute before execute main script.
	// This script should prepare environment for main script (e.g. install software, create config files, build a cluster).
	Before Script
	// (Optional) The main script.
	Main Script
	// (Optional) It will execute after executed main script.
	// This script should execute post processing of main script (e.g. verify the result, collect statistics data, shutdown a cluster).
	After Script
}

// Script represents contents of execution task.
// For example, this script contains shell script, ansible playbook, etc.
type Script interface {
	fmt.Stringer
	// Type returns type name of this script.
	Type() ScriptType
	// SetAttr sets an attribute.
	SetAttr(key, value interface{})
	// GetAttr gets a value from attributes.
	GetAttr(key interface{}) interface{}
	// Type specific methods
	// ...
}

// ScriptResult represents an execution result of script.
type ScriptResult interface {
	fmt.Stringer
	StartTime() time.Time
	EndTime() time.Time
	// Host returns host ID where the script was executed.
	Host() string
	// Output returns byte slice of script output.
	Output() []byte
	// ExitCode returns the exit code returned by script process.
	// If the result have not exit code, it should return the NO_EXIT_CODE.
	ExitCode() int
}

// ScriptExecutor implements how execute script of specified type.
// Some executors are provides by official.  See github.com/yuuki0xff/clustertest/executors package.
// Depending on the provisioner, the implementation of executor may be changed or wrapped.
type ScriptExecutor interface {
	fmt.Stringer
	// Type returns supported script type.
	Type() ScriptType
	// Execute method executes the script and returns result.
	// If unsupported type of script passed, it will be panic.
	// Caller must check script type before calling this method.
	Execute(script Script) ScriptResult
}

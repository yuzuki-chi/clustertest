package models

// Provisioner build/manage/destroy infrastructures.
//
// The infrastructure specification called to Spec.
// Spec is specified when creating a Provisioner instance.
type Provisioner interface {
	Create() error
	Delete() error
	Spec() Spec
	Config() InfraConfig
	ScriptSet() *ScriptSet
	ScriptExecutor(scriptType ScriptType) ScriptExecutor
}

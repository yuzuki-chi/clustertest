package proxmoxve

import (
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/clustertest/provisioners"
)

func init() {
	provisioners.Provisioners[models.SpecType("proxomox-ve")] = func(spec models.Spec) models.Provisioner {
		return &PveProvisioner{
			spec: spec.(*PveSpec),
		}
	}
}

type PveProvisioner struct {
	spec   *PveSpec
	config models.InfraConfig // 具体的な型を入れる
}

func (s *PveProvisioner) Create() error {
	// todo: get token
	// todo: create resources
	// todo: check resource status
	// todo: update infra config
}
func (s *PveProvisioner) Delete() error {
	// todo: get token
	// todo: delete resources
	// todo: check resource status
	// todo: update infra config

}
func (s *PveProvisioner) Spec() models.Spec {
	return s.spec
}
func (s *PveProvisioner) Config() models.InfraConfig {
	// todo
}
func (s *PveProvisioner) ScriptExecutor(scriptType models.ScriptType) models.ScriptExecutor {
	switch scriptType {
	case models.ScriptType("remote-exec"):
		// todo: use default impl
	case models.ScriptType("local-exec"):
		// todo: use default impl
	default:
		// todo: not implemented
	}
}
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

func (p *PveProvisioner) Create() error {
	// todo: get client
	// todo: create resources
	// todo: check resource status
	// todo: update infra config
	panic("not implemented")
}
func (p *PveProvisioner) Delete() error {
	// todo: get client
	// todo: delete resources
	// todo: check resource status
	// todo: update infra config
	panic("not implemented")

}
func (p *PveProvisioner) Spec() models.Spec {
	return p.spec
}
func (p *PveProvisioner) Config() models.InfraConfig {
	// todo
	panic("not implemented")
}
func (p *PveProvisioner) ScriptExecutor(scriptType models.ScriptType) models.ScriptExecutor {
	switch scriptType {
	case models.ScriptType("remote-exec"):
		// todo: use default impl
	case models.ScriptType("local-exec"):
		// todo: use default impl
	default:
		// todo: not implemented
	}
	panic("not implemented")
}
func (p *PveProvisioner) client() *PveClient {
	px := p.spec.Proxmox
	return &PveClient{
		Address:     px.Address,
		User:        px.Account.User,
		Password:    px.Account.Password,
		Fingerprint: px.Fingerprint,
	}
}

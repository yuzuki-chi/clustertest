package proxmoxve

import (
	"github.com/pkg/errors"
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

// Create creates all resources of defined by PveSpec.
func (p *PveProvisioner) Create() error {
	c := p.client()
	err := c.Ticket()
	if err != nil {
		return errors.Wrap(err, "failed to get Proxmox VE API ticket")
	}

	// Create resources.
	for _, vm := range p.spec.VMs {
		from, err := c.IDFromName(vm.Template)
		if err != nil {
			return errors.Wrapf(err, "not found template: %s", vm.Template)
		}
		for i := 0; i < vm.Nodes; i++ {
			// TODO: get an address from address pools.
			ip := ""
			// TODO: decide node.
			nodeID := NodeID("")

			toVMID, err := c.RandomVMID()
			if err != nil {
				return errors.Wrap(err, "failed to generate a random id")
			}
			to := NodeVMID{
				NodeID: nodeID,
				VMID:   toVMID,
			}
			err = c.CloneVM(from, to, "", "This VM created by clustertest-proxmox-ve-provisioner")
			if err != nil {
				return errors.Wrap(err, "failed to clone")
			}
			if vm.StorageSize > 0 {
				err = c.ResizeVolume(to, "scsi0", vm.StorageSize)
				if err != nil {
					return errors.Wrap(err, "failed to resize")
				}
			}
			err = c.UpdateConfig(to, &Config{
				CPUCores:   vm.Processors,
				CPUSockets: 1,
				Memory:     vm.MemorySize * 1024,
				User:       p.spec.User.User,
				SSHKeys:    p.spec.User.SSHPublicKey,
				IPAddress:  ip,
			})
			if err != nil {
				return errors.Wrap(err, "failed to update config")
			}
		}
	}

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

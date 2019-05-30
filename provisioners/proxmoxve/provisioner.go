package proxmoxve

import (
	"github.com/pkg/errors"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/clustertest/provisioners"
	"github.com/yuuki0xff/clustertest/provisioners/proxmoxve/addresspool"
	"net"
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

	err = p.updateSchedulerStatus(c)
	if err != nil {
		return errors.Wrap(err, "failed to update global scheduler status")
	}

	segs, err := p.segments()
	if err != nil {
		return err
	}

	// Create resources.
	err = GlobalScheduler.Transaction(func(scheduler *ScheduleTx) error {
		return addresspool.GlobalPool.Transaction(func(pool *addresspool.AddressPoolTx) error {
			for _, vm := range p.spec.VMs {
				from, err := c.IDFromName(vm.Template)
				if err != nil {
					return errors.Wrapf(err, "not found template: %s", vm.Template)
				}
				for i := 0; i < vm.Nodes; i++ {
					// Allocate resources.
					ip := pool.Allocate(segs)
					nodeID, err := scheduler.Schedule(VMSpec{
						Processors: vm.Processors,
						Memory:     vm.MemorySize,
					})
					if err != nil {
						return err
					}

					// Generate Random ID
					toVMID, err := c.RandomVMID()
					if err != nil {
						return errors.Wrap(err, "failed to generate a random id")
					}
					to := NodeVMID{
						NodeID: nodeID,
						VMID:   toVMID,
					}

					// Clone specified VM and set up it.
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
			return nil
		})
	})
	if err != nil {
		return err
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
func (p *PveProvisioner) segments() ([]addresspool.Segment, error) {
	var segs []addresspool.Segment
	for _, pconf := range p.spec.AddressPools {
		start := net.ParseIP(pconf.StartAddress)
		end := net.ParseIP(pconf.EndAddress)
		gateway := net.ParseIP(pconf.Gateway)
		if start == nil {
			return nil, errors.Errorf("the StartAddress is invalid address: %s", pconf.StartAddress)
		}
		if end == nil {
			return nil, errors.Errorf("the EndAddress is invalid address: %s", pconf.EndAddress)
		}
		if gateway == nil {
			return nil, errors.Errorf("the Gateway is invalid address: %s", pconf.Gateway)
		}

		segs = append(segs, addresspool.Segment{
			StartAddress: start,
			EndAddress:   end,
			Mask:         uint(pconf.CIDR),
			Gateway:      gateway,
		})
	}
	return segs, nil
}
func (p *PveProvisioner) updateSchedulerStatus(c *PveClient) error {
	return GlobalScheduler.UpdateNodes(func() ([]*Node, error) {
		nodeInfos, err := c.ListNodes()
		if err != nil {
			return nil, err
		}

		var nodes []*Node
		for _, n := range nodeInfos {
			var totalCPUs int
			var totalMem int
			vms, err := c.ListVMs(n.ID)
			if err != nil {
				return nil, err
			}
			for _, vm := range vms {
				totalCPUs += vm.Cpus
				totalMem += vm.Mem
			}

			nodes = append(nodes, &Node{
				NodeID: n.ID,
				PCPU:   n.MaxCPU,
				VCPU: struct {
					Max      int
					Used     int
					Reserved int
				}{Max: 0, Used: totalCPUs, Reserved: 0},
				PMem: n.MaxMem,
				VMem: struct {
					System   int
					Used     int
					Reserved int
				}{System: DEFAULT_SYSTEM_MEM, Used: totalMem, Reserved: 0},
			})
		}
		return nodes, nil
	}, false)
}

package proxmoxve

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/clustertest/provisioners"
	"github.com/yuuki0xff/clustertest/provisioners/proxmoxve/addresspool"
	. "github.com/yuuki0xff/clustertest/provisioners/proxmoxve/api"
	"net"
	"time"
)

const CloneTimeout = 30 * time.Second

func init() {
	provisioners.Provisioners[models.SpecType("proxmox-ve")] = func(spec models.Spec) models.Provisioner {
		return &PveProvisioner{
			spec: spec.(*PveSpec),
		}
	}
}

type PveProvisioner struct {
	spec   *PveSpec
	config *PveInfraConfig
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
	conf := NewPveInfraConfig(p.spec)
	err = GlobalScheduler.Transaction(func(scheduler *ScheduleTx) error {
		return addresspool.GlobalPool.Transaction(func(pool *addresspool.AddressPoolTx) error {
			for vmGroupName, vm := range p.spec.VMs {
				from, err := c.IDFromName(vm.Template)
				if err != nil {
					return errors.Wrapf(err, "not found template: %s", vm.Template)
				}
				for i := 0; i < vm.Nodes; i++ {
					// Allocate resources.
					s, ip, ok := pool.AllocateIP(segs)
					if !ok {
						return errors.Errorf("failed to allocate IP address")
					}
					vmSpec := VMSpec{
						Processors: vm.Processors,
						Memory:     vm.MemorySize,
					}
					nodeID, err := scheduler.Schedule(vmSpec)
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

					conf.AddVM(vmGroupName, VMConfig{
						ID:   to,
						IP:   ip,
						Spec: vmSpec,
					})

					// Clone specified VM and set up it.
					vmName := fmt.Sprintf("%s-%s-%d", p.spec.Name, vmGroupName, i)
					task, err := c.CloneVM(from, to, vmName, "This VM created by clustertest-proxmox-ve-provisioner")
					if err != nil {
						return errors.Wrap(err, "failed to clone")
					}

					// Wait for clone operation to complete.
					ctx, _ := context.WithTimeout(context.Background(), CloneTimeout)
					err = task.Wait(ctx)
					if err != nil {
						return errors.Wrap(err, "clone operation is timeout")
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
						IPAddress:  addresspool.ToPveIPConf(s, ip),
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
		// TODO: remove allocated resources
		return err
	}
	// Update the InfraConfig.
	p.config = conf

	// Check resource status.
	for _, vms := range conf.VMs {
		for _, vm := range vms {
			info, err := c.VMInfo(vm.ID)
			if err != nil {
				return err
			}
			if info.Status != "stopped" {
				return fmt.Errorf("invalid status: %s (id=%s)", info.Status, vm.ID)
			}
			// OK
		}
	}
	return nil
}

// Delete deletes all resources of defined by PveSpec.
func (p *PveProvisioner) Delete() error {
	if p.config == nil {
		return errors.Errorf("still not provisioned")
	}

	c := p.client()
	err := c.Ticket()
	if err != nil {
		return err
	}

	// Delete resources.
	for _, vms := range p.config.VMs {
		for _, vm := range vms {
			err := c.DeleteVM(vm.ID)
			if err != nil {
				return errors.Wrap(err, "failed to delete VM")
			}
			addresspool.GlobalPool.Free(vm.IP)
			GlobalScheduler.Free(vm.ID.NodeID, vm.Spec)
		}
	}

	// All resources are deleted.
	// Should discard the InfraConfig.
	p.config = nil
	return nil
}
func (p *PveProvisioner) Spec() models.Spec {
	return p.spec
}
func (p *PveProvisioner) Config() models.InfraConfig {
	return p.config
}
func (p *PveProvisioner) ScriptSet() *models.ScriptSet {
	var sets []*models.ScriptSet
	for _, vmGroup := range p.spec.VMs {
		s := &models.ScriptSet{
			Before: vmGroup.Scripts.Before.Get(),
			Main:   vmGroup.Scripts.Main.Get(),
			After:  vmGroup.Scripts.After.Get(),
		}
		sets = append(sets, s)
	}

	// Merge multiple ScriptSet to a single ScriptSet.
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
				if vm.Status == "running" {
					totalCPUs += vm.Cpus
					totalMem += vm.Mem
				}
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

package proxmoxve

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/yuuki0xff/clustertest/executors"
	"github.com/yuuki0xff/clustertest/executors/callback"
	"github.com/yuuki0xff/clustertest/executors/localshell"
	"github.com/yuuki0xff/clustertest/executors/remoteshell"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/clustertest/provisioners"
	"github.com/yuuki0xff/clustertest/provisioners/proxmoxve/addresspool"
	. "github.com/yuuki0xff/clustertest/provisioners/proxmoxve/api"
	"golang.org/x/sync/errgroup"
	"net"
	"sync"
	"time"
)

const CloneTimeout = 30 * time.Second
const StartTimeout = 10 * time.Second
const DeleteTimeout = 10 * time.Second

const specType = models.SpecType("proxmox-ve")
const vmConfigsAttrName = "provisioners/proxmox-ve/vm-configs"

// TODO
// タスクが動いている最中、特にReserve()とCreate()の間にschedulerStatusが実行されてしまうと、
// 確保済み扱いのCPUやメモリがカウントされない。このため、リソース開放時にpanicする。
// これを回避するために Scheduler.UpdateNodes() の実行は1階限りに制限する。
var PveUpdateSchedulerOnce sync.Once

func init() {
	provisioners.Provisioners[specType] = func(prefix string, spec models.Spec) models.Provisioner {
		return &PveProvisioner{
			prefix: prefix,
			spec:   spec.(*PveSpec),
		}
	}
}

type PveProvisioner struct {
	// Prefix for all VM names
	prefix string
	spec   *PveSpec
	config *PveInfraConfig
}

// Reserve() reserves all resources (CPU, memory, storage, etc) of defined by PveSpec.
func (p *PveProvisioner) Reserve() error {
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
			eg := errgroup.Group{}
			for vmGroupName, vm := range p.spec.VMs {
				from, err := c.IDFromName(vm.Template)
				if err != nil {
					return errors.Wrapf(err, "not found template: %s", vm.Template)
				}
				for i := 0; i < vm.Nodes; i++ {
					err := p.allocateVM(c, conf, &eg, segs, scheduler, pool, vmGroupName, vm, from, i)
					if err != nil {
						return err
					}
				}
			}
			return eg.Wait()
		})
	})
	if err != nil {
		// TODO: remove allocated resources
		return err
	}
	// Update the InfraConfig.
	p.config = conf
	return nil
}

// Create starts all VMs of defined by PveSpec.
func (p *PveProvisioner) Create() error {
	c := p.client()
	err := c.Ticket()
	if err != nil {
		return errors.Wrap(err, "failed to get Proxmox VE API ticket")
	}

	// Start all VMs.
	eg := errgroup.Group{}
	conf := p.config
	for _, vms := range conf.VMs {
		for _, vm := range vms {
			vm := vm
			eg.Go(func() error {
				task, err := c.StartVM(vm.ID)
				if err != nil {
					return err
				}
				ctx, _ := context.WithTimeout(context.Background(), StartTimeout)
				return task.Wait(ctx)
			})
		}
	}
	err = eg.Wait()
	if err != nil {
		return err
	}

	// Check resource status.
	eg = errgroup.Group{}
	for _, vms := range conf.VMs {
		for _, vm := range vms {
			vm := vm
			eg.Go(func() error {
				info, err := c.VMInfo(vm.ID)
				if err != nil {
					return err
				}
				if info.Status != RunningVMStatus {
					return fmt.Errorf("invalid status: %s (id=%s)", info.Status, vm.ID)
				}
				// OK
				return nil
			})
		}
	}
	return eg.Wait()
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
	eg := errgroup.Group{}
	for _, vms := range p.config.VMs {
		for _, vm := range vms {
			vm := vm
			eg.Go(func() error {
				task, err := c.DeleteVM(vm.ID)
				if err != nil {
					return errors.Wrap(err, "failed to delete VM")
				}
				ctx, _ := context.WithTimeout(context.Background(), DeleteTimeout)
				err = task.Wait(ctx)
				if err != nil {
					return err
				}

				addresspool.GlobalPool.Free(vm.IP)
				GlobalScheduler.Free(vm.ID.NodeID, vm.Spec)
				return nil
			})
		}
	}
	err = eg.Wait()
	if err != nil {
		return nil
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
func (p *PveProvisioner) ScriptSets() []*models.ScriptSet {
	var sets []*models.ScriptSet
	for name, vmGroup := range p.spec.VMs {
		attrs := map[interface{}]interface{}{
			vmConfigsAttrName: p.config.VMs[name],
		}
		s := &models.ScriptSet{
			Before: vmGroup.Scripts.Before.SetAttrs(attrs).Get(),
			Main:   vmGroup.Scripts.Main.SetAttrs(attrs).Get(),
			After:  vmGroup.Scripts.After.SetAttrs(attrs).Get(),
		}
		sets = append(sets, s)
	}
	return sets
}
func (p *PveProvisioner) ScriptExecutor(scriptType models.ScriptType) models.ScriptExecutor {
	var newExecutor func(config *VMConfig, script models.Script) models.ScriptExecutor

	switch scriptType {
	case models.ScriptType("remote-shell"):
		newExecutor = func(config *VMConfig, script models.Script) models.ScriptExecutor {
			return &remoteshell.Executor{
				User: p.spec.User.User,
				Host: config.IP.String(),
			}
		}
	case models.ScriptType("local-shell"):
		newExecutor = func(config *VMConfig, script models.Script) models.ScriptExecutor {
			return &localshell.Executor{}
		}
	default:
		err := errors.Errorf("unsupported ScriptType: %s", scriptType)
		panic(err)
	}

	return &callback.Executor{
		Fn: func(script models.Script) models.ScriptResult {
			mr := &executors.MergedResult{}
			vmConfigs := script.GetAttr(vmConfigsAttrName).([]VMConfig)
			for _, c := range vmConfigs {
				e := newExecutor(&c, script)
				mr.Append(e.Execute(script))
			}
			return mr
		},
	}
}
func (p *PveProvisioner) client() *PveClient {
	px := p.spec.Proxmox
	return NewPveClient(PveClientOption{
		Address:     px.Address,
		User:        px.Account.User,
		Password:    px.Account.Password,
		Fingerprint: px.Fingerprint,
	})
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
func (p *PveProvisioner) updateSchedulerStatus(c *PveClient) (err error) {
	PveUpdateSchedulerOnce.Do(func() {
		err = GlobalScheduler.UpdateNodes(func() ([]*Node, error) {
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
					PMem: byte2megabyte(n.MaxMem),
					VMem: struct {
						System   int
						Used     int
						Reserved int
					}{
						System:   DEFAULT_SYSTEM_MEM,
						Used:     byte2megabyte(totalMem),
						Reserved: 0,
					},
				})
			}
			return nodes, nil
		}, false)
	})
	return
}
func (p *PveProvisioner) allocateVM(
	c *PveClient,
	conf *PveInfraConfig,
	eg *errgroup.Group,
	segs []addresspool.Segment,
	scheduler *ScheduleTx,
	pool *addresspool.AddressPoolTx,
	vmGroupName string,
	vm *PveVM,
	from NodeVMID,
	i int,
) error {
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
	vmName := fmt.Sprintf("%s-%s-%s-%d", p.prefix, p.spec.Name, vmGroupName, i)
	description := fmt.Sprintf(
		"This VM created by clustertest-proxmox-ve-provisioner.\n"+
			"\n"+
			"TaskName: %s\n"+
			"SpecName: %s\n"+
			"GroupName: %s\n"+
			"Index: %d\n"+
			"\n"+
			"Created at %s\n"+
			"IP: %s\n",
		p.prefix,
		p.spec.Name,
		vmGroupName,
		i,
		time.Now().String(),
		ip.String(),
	)
	task, err := c.CloneVM(from, to, vmName, description, vm.Pool)
	if err != nil {
		return errors.Wrap(err, "failed to clone")
	}

	eg.Go(func() error {
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
			Memory:     vm.MemorySize,
			User:       p.spec.User.User,
			SSHKeys:    p.spec.User.SSHPublicKey,
			IPAddress:  addresspool.ToPveIPConf(s, ip),
		})
		if err != nil {
			return errors.Wrap(err, "failed to update config")
		}
		return nil
	})
	return nil
}

// byte2megabyte converts byte to MiB.
func byte2megabyte(b int) int {
	return b / 1024 / 1024
}

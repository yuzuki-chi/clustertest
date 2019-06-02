package proxmoxve

import (
	"github.com/yuuki0xff/clustertest/models"
	. "github.com/yuuki0xff/clustertest/provisioners/proxmoxve/api"
	"net"
	"sync"
)

type PveInfraConfig struct {
	PveSpec *PveSpec
	VMs     map[string][]VMConfig
	m       sync.Mutex
}
type VMConfig struct {
	ID   NodeVMID
	IP   net.IP
	Spec VMSpec
}

func NewPveInfraConfig(spec *PveSpec) *PveInfraConfig {
	return &PveInfraConfig{
		PveSpec: spec,
		VMs:     map[string][]VMConfig{},
	}
}

func (c *PveInfraConfig) String() string {
	return "<PveInfraConfig>"
}
func (c *PveInfraConfig) Spec() models.Spec {
	return c.PveSpec
}
func (c *PveInfraConfig) AddVM(name string, vm VMConfig) {
	c.m.Lock()
	defer c.m.Unlock()
	c.VMs[name] = append(c.VMs[name], vm)
}

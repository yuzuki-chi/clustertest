package proxmoxve

import (
	"github.com/yuuki0xff/clustertest/config"
	"github.com/yuuki0xff/clustertest/models"
)

func init() {
	config.SpecInitializers[models.SpecType("proxmox-ve")] = func() models.Spec { return &PveSpec{} }
}

type PveSpec struct {
	// Identifier of the spec.
	Name string
	// Proxmox VE account settings..
	Proxmox *struct {
		// URL of the Proxmox VE API server.
		// Example: https://pve.local:8006
		Address string
		Account struct {
			User     string
			Password string
		}
		// Fingerprint of the Proxmox VE API server.
		// If you need the server certificate pinning to make it more secure.
		Fingerprint string
	}
	// Addresses to assign to VMs.
	AddressPools []*struct {
		StartAddress string `yaml:"start_address"`
		EndAddress   string `yaml:"end_address"`
		CIDR         int
		Gateway      string
	} `yaml:"address_pools"`
	// User information.
	// This user will create by cloud-init at VM start-up.
	User *struct {
		User         string
		Password     string
		SSHPublicKey string `yaml:"ssh_public_key"`
	}
	VMs map[string]*struct {
		// Template name.
		Template string
		// Number of VMs.
		Nodes int
		// Number of processors.
		Processors int
		// RAM size (MiB).
		MemorySize int `yaml:"memory_size"`
		// Minimal storage size (GiB).
		// The storage may be large than specified size.
		StorageSize int `yaml:"storage_size"`
		// Define tasks to execute on VMs.
		Scripts *config.ScriptConfig
	}
}

func (s *PveSpec) String() string {
	return "<PveSpec>"
}
func (s *PveSpec) Type() models.SpecType {
	return models.SpecType("proxmox-ve")
}

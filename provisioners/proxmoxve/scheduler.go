package proxmoxve

import "sync"

type Node struct {
	NodeID NodeID

	// Number of physical CPUs.
	PCPU int
	// Information of virtual CPUs.
	VCPU struct {
		// Maximum number of virtual CPUs.
		// If you need disable the maximum limit of virtual CPUs, set Max=0.
		Max int
		// Number of used virtual CPUs by running VMs.
		Used int
		// Number of virtual CPUs by VMs in preparing.
		Reserved int
	}

	// Amount of physical memory size (MiB).
	PMem int
	// Information of virtual memory.
	VMem struct {
		// Amount of memory size (MiB) for system.
		// This memory is used by Proxmox VE system and margin for prevent out-of-memory.
		// Default: 4096 (4096MiB = 4GiB)
		System int
		// Amount of used memory size (MiB) by already running VMs.
		// This value is the total of memory size each running VMs.
		Used int
		// Amount of reserved memory size (MiB).
		// This value is the total of memory size each VMs in preparing.
		Reserved int
	}
}
type VMSpec struct {
	// Number of processors
	Processors int
	// RAM size (MiB)
	Memory int
}

type Scheduler struct {
	nodes map[NodeID]*Node
	m     sync.RWMutex
}
type ScheduleTx struct {
	s        *Scheduler
	reserved []struct {
		Spec VMSpec
		ID   NodeID
	}
	committed bool
	reverted bool
}

// UpdateNodes updates all nodes status.
// If updateReserved is true, Scheduler.nodes.VCPU.Reserved and Scheduler.nodes.VMem.Reserved parameters are updated.
func (s *Scheduler) UpdateNodes(fn func() ([]*Node, error), updateReserved bool) error {}

// Schedule decides best VM location and reserves it.
func (s *Scheduler) Schedule(spec VMSpec) (NodeID, error) {}

// Use notifies it to scheduler that specified VM started to running.
func (s *Scheduler) Use(id NodeID, spec VMSpec) {}

// Free releases reserved resources of vCPU and vMem.
// You should call it when you need drops the reserved resources by Schedule().
func (s *Scheduler) Free(id NodeID, spec VMSpec) {}

// Transaction starts an transaction to allocate resources atomically.
func (s *Scheduler) Transaction(fn func(tx *ScheduleTx) error) error {}

// Schedule decides best VM location and reserves it.
// The reserved resources are allocated or released when calling the Commit() or Revert().
func (tx *ScheduleTx) Schedule(spec VMSpec) (NodeID, error) {}

// Commit allocates all reserved resources while this transaction.
func (tx *ScheduleTx) Commit() {}

// Revert releases all reserved resources while this transaction.
func (tx *ScheduleTx) Revert() {}

// IsFinished returns true if already committed or reverted.
// Otherwise, it returns false.
func (tx *ScheduleTx) IsFinished() bool {}

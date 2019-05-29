package proxmoxve

import (
	"github.com/pkg/errors"
	"sync"
)

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
	S        *Scheduler
	m        sync.Mutex
	reserved []struct {
		ID   NodeID
		Spec VMSpec
	}
	committed bool
	reverted  bool
}

// UpdateNodes updates all nodes status.
// If updateReserved is true, Scheduler.nodes.VCPU.Reserved and Scheduler.nodes.VMem.Reserved parameters are updated.
func (s *Scheduler) UpdateNodes(fn func() ([]*Node, error), updateReserved bool) error {
	panic("todo") // TODO
}

// Schedule decides best VM location and reserves it.
func (s *Scheduler) Schedule(spec VMSpec) (NodeID, error) {
	panic("todo") // TODO
}

// Use notifies it to scheduler that specified VM started to running.
func (s *Scheduler) Use(id NodeID, spec VMSpec) {
	s.m.Lock()
	defer s.m.Unlock()

	node := s.nodes[id]
	if node == nil {
		panic(errors.Errorf("not found node: %s", id))
	}
	if node.VCPU.Reserved-spec.Processors < 0 {
		panic(errors.Errorf("lacking reserved vCPUs"))
	}
	if node.VMem.Reserved-spec.Memory < 0 {
		panic(errors.Errorf("lacking reserved vMem"))
	}

	node.VCPU.Reserved -= spec.Processors
	node.VCPU.Used += spec.Processors

	node.VMem.Reserved -= spec.Memory
	node.VMem.Used += spec.Memory
}

// Free releases reserved resources of vCPU and vMem.
// You should call it when you need drops the reserved resources by Schedule().
func (s *Scheduler) Free(id NodeID, spec VMSpec) {
	s.m.Lock()
	defer s.m.Unlock()

	node := s.nodes[id]
	if node == nil {
		panic(errors.Errorf("not found node: %s", id))
	}
	if node.VCPU.Reserved-spec.Processors < 0 {
		panic(errors.Errorf("lacking reserved vCPUs"))
	}
	if node.VMem.Reserved-spec.Memory < 0 {
		panic(errors.Errorf("lacking reserved vMem"))
	}

	node.VCPU.Reserved -= spec.Processors
	node.VMem.Reserved -= spec.Memory
}

// Transaction starts an transaction to allocate resources atomically.
func (s *Scheduler) Transaction(fn func(tx *ScheduleTx) error) error {
	tx := ScheduleTx{S: s}

	err := fn(&tx)
	if tx.IsFinished() {
		// No need to commit or revert.
		return err
	}

	if err != nil {
		tx.Revert()
	} else {
		tx.Commit()
	}
	return err
}

// Schedule decides best VM location and reserves it.
// The reserved resources are allocated or released when calling the Commit() or Revert().
func (tx *ScheduleTx) Schedule(spec VMSpec) (NodeID, error) {
	tx.m.Lock()
	defer tx.m.Unlock()

	tx.mustNotFinished()
	id, err := tx.S.Schedule(spec)
	if err != nil {
		return NodeID(""), err
	}

	tx.reserved = append(tx.reserved, struct {
		ID   NodeID
		Spec VMSpec
	}{ID: id, Spec: spec})
	return id, nil
}

// Commit allocates all reserved resources while this transaction.
func (tx *ScheduleTx) Commit() {
	tx.m.Lock()
	defer tx.m.Unlock()
	tx.mustNotFinished()
	for _, r := range tx.reserved {
		tx.S.Use(r.ID, r.Spec)
	}
}

// Revert releases all reserved resources while this transaction.
func (tx *ScheduleTx) Revert() {
	tx.m.Lock()
	defer tx.m.Unlock()
	tx.mustNotFinished()
	for _, r := range tx.reserved {
		tx.S.Free(r.ID, r.Spec)
	}
}

// IsFinished returns true if already committed or reverted.
// Otherwise, it returns false.
func (tx *ScheduleTx) IsFinished() bool {
	tx.m.Lock()
	defer tx.m.Unlock()
	return tx.committed || tx.reverted
}

// mustNotFinished checks that if whether this transaction finished.
// If it finished, mustNotFinished will panics.
func (tx *ScheduleTx) mustNotFinished() {
	if tx.committed {
		panic(errors.Errorf("this transaction has been committed"))
	}
	if tx.reverted {
		panic(errors.Errorf("this transaction has been reverted"))
	}
}

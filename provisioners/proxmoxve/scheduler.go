package proxmoxve

import (
	"context"
	"github.com/pkg/errors"
	. "github.com/yuuki0xff/clustertest/provisioners/proxmoxve/api"
	"sync"
	"time"
)

// 4096MiB = 4GiB
const DEFAULT_SYSTEM_MEM = 4096

// TODO: 複数のクラスタに対応できない
var GlobalScheduler = &Scheduler{}

// FullError means failed to allocate resources.
var FullError = errors.Errorf("full")

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
		// Default: DEFAULT_SYSTEM_MEM
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
	m     sync.Mutex
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
	s.m.Lock()
	defer s.m.Unlock()

	nodes, err := fn()
	if err != nil {
		return err
	}
	m := map[NodeID]*Node{}
	for _, n := range nodes {
		if !updateReserved {
			oldNode := s.nodes[n.NodeID]
			if oldNode != nil {
				// Copy the Reserved parameter values from old data.
				newNode := &Node{}
				*newNode = *n
				newNode.VCPU.Reserved = oldNode.VCPU.Reserved
				newNode.VMem.Reserved = oldNode.VMem.Reserved
				n = newNode
			} else {
				// Old data is not found.
				// Treat the Reserved parameter as 0.
				newNode := &Node{}
				*newNode = *n
				newNode.VCPU.Reserved = 0
				newNode.VMem.Reserved = 0
				n = newNode
			}
		} else {
			newNode := &Node{}
			*newNode = *n
			n = newNode
		}

		if n.VMem.System == 0 {
			n.VMem.System = DEFAULT_SYSTEM_MEM
		}
		m[n.NodeID] = n
	}
	s.nodes = m
	return nil
}

// Schedule decides best VM location and reserves it.
// You should call the Use() or Cancel() after call it.
func (s *Scheduler) Schedule(spec VMSpec) (NodeID, error) {
	s.m.Lock()
	defer s.m.Unlock()

	for id, n := range s.nodes {
		if n.VCPU.Max > 0 {
			if n.VCPU.Max-(n.VCPU.Used+n.VCPU.Reserved) < spec.Processors {
				// insufficient vCPUs exists.
				continue
			}
		}
		if n.PMem-(n.VMem.System+n.VMem.Used+n.VMem.Reserved) < spec.Memory {
			// insufficient memory exists.
			continue
		}
		// Found a best node.
		n.VCPU.Reserved += spec.Processors
		n.VMem.Reserved += spec.Memory
		return id, nil
	}
	// Not found a best node.
	return NodeID(""), FullError
}

// Use notifies it to scheduler that specified VM started to running.
// We regards the reserved resources as in use.
// You should call Free() when release resources.
func (s *Scheduler) Use(id NodeID, spec VMSpec) {
	s.m.Lock()
	defer s.m.Unlock()

	node := s.nodes[id]
	if node == nil {
		panic(errors.Errorf("not found node: %s", id))
	}
	if node.VCPU.Reserved-spec.Processors < 0 {
		panic(errors.Errorf("lacking reserved vCPUs: %d %d", node.VCPU.Reserved, spec.Processors))
	}
	if node.VMem.Reserved-spec.Memory < 0 {
		panic(errors.Errorf("lacking reserved vMem: %d %d", node.VMem.Reserved, spec.Memory))
	}

	node.VCPU.Reserved -= spec.Processors
	node.VCPU.Used += spec.Processors

	node.VMem.Reserved -= spec.Memory
	node.VMem.Used += spec.Memory
}

// Cancel releases reserved resources.
// You should call it when you need drops the reserved resources by Schedule().
func (s *Scheduler) Cancel(id NodeID, spec VMSpec) {
	s.m.Lock()
	defer s.m.Unlock()

	node := s.nodes[id]
	if node == nil {
		panic(errors.Errorf("not found node: %s", id))
	}
	if node.VCPU.Reserved-spec.Processors < 0 {
		panic(errors.Errorf("lacking reserved vCPUs: %d %d", node.VCPU.Reserved, spec.Processors))
	}
	if node.VMem.Reserved-spec.Memory < 0 {
		panic(errors.Errorf("lacking reserved vMem: %d %d", node.VMem.Reserved, spec.Memory))
	}

	node.VCPU.Reserved -= spec.Processors
	node.VMem.Reserved -= spec.Memory
}

func (s *Scheduler) Free(id NodeID, spec VMSpec) {
	s.m.Lock()
	defer s.m.Unlock()

	node := s.nodes[id]
	if node == nil {
		panic(errors.Errorf("not found node: %s", id))
	}
	if node.VCPU.Used-spec.Processors < 0 {
		panic(errors.Errorf("lacking reserved vCPUs: %d %d", node.VCPU.Used, spec.Processors))
	}
	if node.VMem.Used-spec.Memory < 0 {
		panic(errors.Errorf("lacking reserved vMem: %d %d", node.VMem.Used, spec.Memory))
	}

	node.VCPU.Used -= spec.Processors
	node.VMem.Used -= spec.Memory
}

// Transaction starts an transaction to allocate resources atomically.
func (s *Scheduler) Transaction(fn func(tx *ScheduleTx) error) error {
	// NOTE: MUST NOT acquire the lock on this method.
	// Because Schedule(), Use() and Free() are called some times while running the fn(),

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

// ScheduleWait waits for allocate resources and reserves it.
func (tx *ScheduleTx) ScheduleWait(ctx context.Context, spec VMSpec) (NodeID, error) {
	t := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return NodeID(""), errors.New("timeout ScheduleWait()")
		case <-t.C:
			id, err := tx.Schedule(spec)
			if err == FullError {
				continue
			}
			return id, err
		}
	}
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
		tx.S.Cancel(r.ID, r.Spec)
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

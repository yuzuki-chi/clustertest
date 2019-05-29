package proxmoxve

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScheduler_UpdateNodes(t *testing.T) {
	t.Run("should_keep_reserved_parameters_of_exists_nodes", func(t *testing.T) {
		initNodes := map[NodeID]*Node{
			NodeID("1"): {
				NodeID: NodeID("1"),
				PCPU:   1,
				PMem:   2,
				VCPU: struct {
					Max      int
					Used     int
					Reserved int
				}{Max: 3, Used: 4, Reserved: 111},
				VMem: struct {
					System   int
					Used     int
					Reserved int
				}{System: 6, Used: 7, Reserved: 222},
			},
		}
		node := &Node{
			NodeID: NodeID("1"),
			PCPU:   1,
			PMem:   2,
			VCPU: struct {
				Max      int
				Used     int
				Reserved int
			}{Max: 3, Used: 4, Reserved: 5},
			VMem: struct {
				System   int
				Used     int
				Reserved int
			}{System: 6, Used: 7, Reserved: 8},
		}

		s := Scheduler{
			nodes: initNodes,
		}
		err := s.UpdateNodes(func() ([]*Node, error) {
			return []*Node{node}, nil
		}, false)
		assert.NoError(t, err)
		if !assert.Len(t, s.nodes, 1) {
			return
		}
		assert.Equal(t, 111, s.nodes[NodeID("1")].VCPU.Reserved)
		assert.Equal(t, 222, s.nodes[NodeID("1")].VMem.Reserved)
	})

	t.Run("should_set_0_to_reserved_parameters_of_not_exists_nodes", func(t *testing.T) {
		node := &Node{
			NodeID: NodeID("1"),
			PCPU:   1,
			PMem:   2,
			VCPU: struct {
				Max      int
				Used     int
				Reserved int
			}{Max: 3, Used: 4, Reserved: 111},
			VMem: struct {
				System   int
				Used     int
				Reserved int
			}{System: 6, Used: 7, Reserved: 222},
		}

		s := Scheduler{}
		err := s.UpdateNodes(func() ([]*Node, error) {
			return []*Node{node}, nil
		}, false)
		assert.NoError(t, err)
		if !assert.Len(t, s.nodes, 1) {
			return
		}
		assert.Equal(t, 0, s.nodes[NodeID("1")].VCPU.Reserved)
		assert.Equal(t, 0, s.nodes[NodeID("1")].VMem.Reserved)
	})

	t.Run("should_update_reserved_parameters", func(t *testing.T) {
		node := &Node{
			NodeID: NodeID("1"),
			PCPU:   1,
			PMem:   2,
			VCPU: struct {
				Max      int
				Used     int
				Reserved int
			}{Max: 3, Used: 4, Reserved: 111},
			VMem: struct {
				System   int
				Used     int
				Reserved int
			}{System: 6, Used: 7, Reserved: 222},
		}

		s := Scheduler{}
		err := s.UpdateNodes(func() ([]*Node, error) {
			return []*Node{node}, nil
		}, true)
		assert.NoError(t, err)
		if !assert.Len(t, s.nodes, 1) {
			return
		}
		assert.Equal(t, node, s.nodes[NodeID("1")])
	})
}

func TestScheduler_Schedule(t *testing.T) {
	t.Run("should_not_allocate_on_low_memory_node", func(t *testing.T) {
		s := Scheduler{}
		err := s.UpdateNodes(func() ([]*Node, error) {
			return []*Node{
				// Low memory (free memory is 500 MiB).
				{
					NodeID: NodeID("1"),
					PMem:   1024,
					VMem: struct {
						System   int
						Used     int
						Reserved int
					}{System: 24, Used: 500, Reserved: 0},
				},
			}, nil
		}, false)
		if !assert.NoError(t, err) {
			return
		}

		_, err = s.Schedule(VMSpec{
			Processors: 10,
			Memory:     1000,
		})
		assert.Error(t, err)
	})

	t.Run("should_not_allocate_on_low_cpu_node", func(t *testing.T) {
		s := Scheduler{}
		err := s.UpdateNodes(func() ([]*Node, error) {
			return []*Node{
				// Low vCPUs (Number of free vCPUs is 8).
				{
					NodeID: NodeID("1"),
					VCPU: struct {
						Max      int
						Used     int
						Reserved int
					}{Max: 16, Used: 8, Reserved: 0},
					PMem: 2024,
					VMem: struct {
						System   int
						Used     int
						Reserved int
					}{System: 24, Used: 500, Reserved: 0},
				},
			}, nil
		}, false)
		if !assert.NoError(t, err) {
			return
		}

		_, err = s.Schedule(VMSpec{
			Processors: 10,
			Memory:     1000,
		})
		assert.Error(t, err)
	})

	t.Run("should_allocate_on_free_node", func(t *testing.T) {
		s := Scheduler{}
		err := s.UpdateNodes(func() ([]*Node, error) {
			return []*Node{
				// Free memory is 1000 MiB.
				{
					NodeID: NodeID("1"),
					PMem:   2024,
					VMem: struct {
						System   int
						Used     int
						Reserved int
					}{System: 24, Used: 500, Reserved: 0},
				},
			}, nil
		}, false)
		if !assert.NoError(t, err) {
			return
		}

		id, err := s.Schedule(VMSpec{
			Processors: 10,
			Memory:     1000,
		})
		assert.NoError(t, err)
		assert.Equal(t, NodeID("1"), id)
	})
}

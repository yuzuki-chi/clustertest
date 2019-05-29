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

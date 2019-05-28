package addresspool

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestSegment_Init(t *testing.T) {
	t.Run("should_success_when_valid_args", func(t *testing.T) {
		assert.NotPanics(t, func() {
			pool := &Segment{
				StartAddress: net.IPv4(10, 0, 0, 10),
				EndAddress:   net.IPv4(10, 0, 0, 20),
				Mask:         24,
				Gateway:      net.IPv4(10, 0, 0, 1),
			}
			pool.Init()
		})
	})

	t.Run("should_fail_when_gateway_address_is_not_valid", func(t *testing.T) {
		assert.Panics(t, func() {
			pool := &Segment{
				StartAddress: net.IPv4(10, 1, 2, 10),
				EndAddress:   net.IPv4(10, 1, 2, 20),
				Mask:         24,
				Gateway:      net.IPv4(10, 0, 0, 1),
			}
			pool.Init()
		})
	})

	t.Run("should_fail_when_end_address_is_not_valid", func(t *testing.T) {
		assert.Panics(t, func() {
			pool := &Segment{
				StartAddress: net.IPv4(10, 0, 0, 10),
				EndAddress:   net.IPv4(10, 1, 2, 20),
				Mask:         24,
				Gateway:      net.IPv4(10, 0, 0, 1),
			}
			pool.Init()
		})
	})
}

func TestSegment_Allocate(t *testing.T) {
	pool := &Segment{
		StartAddress: net.IPv4(10, 0, 0, 10),
		EndAddress:   net.IPv4(10, 0, 0, 12),
		Mask:         24,
		Gateway:      net.IPv4(10, 0, 0, 1),
	}
	pool.Init()
	assert.Equal(t, "gw=10.0.0.1,ip=10.0.0.10/24", pool.Allocate())
	assert.Equal(t, "gw=10.0.0.1,ip=10.0.0.11/24", pool.Allocate())
	assert.Equal(t, "gw=10.0.0.1,ip=10.0.0.12/24", pool.Allocate())
	// full
	assert.Equal(t, "", pool.Allocate())
	assert.Equal(t, "", pool.Allocate())
	// free
	pool.Free(net.IPv4(10, 0, 0, 11))
	assert.Equal(t, "gw=10.0.0.1,ip=10.0.0.11/24", pool.Allocate())
	//full
	assert.Equal(t, "", pool.Allocate())
}

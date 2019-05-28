package addresspool

import (
	"github.com/magiconair/properties/assert"
	"net"
	"testing"
)

func TestAddressPool_Allocate(t *testing.T) {
	segs := []Segment{
		{
			StartAddress: net.IPv4(10, 0, 0, 10),
			EndAddress:   net.IPv4(10, 0, 0, 12),
			Mask:         24,
			Gateway:      net.IPv4(10, 0, 0, 1),
		},
	}
	pool := NewAddressPool("test_allocate")

	assert.Equal(t, "gw=10.0.0.1,ip=10.0.0.10/24", pool.Allocate(segs))
	assert.Equal(t, "gw=10.0.0.1,ip=10.0.0.11/24", pool.Allocate(segs))
	assert.Equal(t, "gw=10.0.0.1,ip=10.0.0.12/24", pool.Allocate(segs))
	// full
	assert.Equal(t, "", pool.Allocate(segs))
	assert.Equal(t, "", pool.Allocate(segs))
	// free
	pool.Free(net.IPv4(10, 0, 0, 11))
	assert.Equal(t, "gw=10.0.0.1,ip=10.0.0.11/24", pool.Allocate(segs))
	//full
	assert.Equal(t, "", pool.Allocate(segs))
}

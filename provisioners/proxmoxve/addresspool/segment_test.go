package addresspool

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestSegmentCursor_Init(t *testing.T) {
	t.Run("should_success_when_valid_args", func(t *testing.T) {
		assert.NotPanics(t, func() {
			s := Segment{
				StartAddress: net.IPv4(10, 0, 0, 10),
				EndAddress:   net.IPv4(10, 0, 0, 20),
				Mask:         24,
				Gateway:      net.IPv4(10, 0, 0, 1),
			}
			s.Validate()
		})
	})

	t.Run("should_fail_when_gateway_address_is_not_valid", func(t *testing.T) {
		assert.Panics(t, func() {
			s := Segment{
				StartAddress: net.IPv4(10, 1, 2, 10),
				EndAddress:   net.IPv4(10, 1, 2, 20),
				Mask:         24,
				Gateway:      net.IPv4(10, 0, 0, 1),
			}
			s.Validate()
		})
	})

	t.Run("should_fail_when_end_address_is_not_valid", func(t *testing.T) {
		assert.Panics(t, func() {
			s := Segment{
				StartAddress: net.IPv4(10, 0, 0, 10),
				EndAddress:   net.IPv4(10, 1, 2, 20),
				Mask:         24,
				Gateway:      net.IPv4(10, 0, 0, 1),
			}
			s.Validate()
		})
	})
}

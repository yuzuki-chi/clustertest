package addresspool

import (
	"github.com/pkg/errors"
	"net"
)

type Segment struct {
	StartAddress net.IP
	EndAddress   net.IP
	Mask         uint
	Gateway      net.IP
}

func (s Segment) Validate() {
	network := s.ip2ipv4(s.StartAddress).network
	if !network.Contains(s.Gateway) {
		panic(errors.Errorf("the Gateway address is out of network: gw=%s network=%s", s.Gateway, network))
	}
	if !network.Contains(s.EndAddress) {
		panic(errors.Errorf("the EndAddress is out of network: gw=%s network=%s", s.Gateway, network))
	}
}

// FindFreeAddress finds non-allocated address and returns it.
func (s Segment) FindFreeAddress(isUsed func(ip net.IP) bool) net.IP {
	endIP := s.ip2ipv4(s.EndAddress)
	cursor := s.ip2ipv4(s.StartAddress)
	for {
		isSpecialAddr := cursor.IsNetwork() || cursor.IsBroadcast()
		if !isSpecialAddr && !isUsed(cursor.ip) {
			// It is usable address.
			return cursor.ip
		}

		// Update cursor position.
		next, err := cursor.Increase()
		if err != nil {
			panic(err)
		}
		cursor = next

		if endIP.LargerThan(cursor) {
			// All addresses are allocated.
			// We can not find the free address.
			return nil
		}
	}
	panic("unreachable")
}
func (s Segment) ip2ipv4(ip net.IP) *IPv4Address {
	return newIPv4AddressByIP(
		ip,
		&net.IPNet{
			IP:   s.StartAddress,
			Mask: net.CIDRMask(int(s.Mask), 32),
		},
	)
}

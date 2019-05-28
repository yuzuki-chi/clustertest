package addresspool

import (
	"fmt"
	"github.com/pkg/errors"
	"net"
)

type Segment struct {
	StartAddress net.IP
	EndAddress   net.IP
	Mask         uint
	Gateway      net.IP

	ipCursor  *IPv4Address
	allocated map[string]struct{}
}

func (p *Segment) Init() {
	network := p.ip2ipv4(p.StartAddress).network
	if !network.Contains(p.Gateway) {
		panic(errors.Errorf("the Gateway address is out of network: gw=%s network=%s", p.Gateway, network))
	}
	if !network.Contains(p.EndAddress) {
		panic(errors.Errorf("the EndAddress is out of network: gw=%s network=%s", p.Gateway, network))
	}

	p.ipCursor = p.ip2ipv4(p.StartAddress)
	p.allocated = map[string]struct{}{}
}

// Allocate allocates an IPv4 address and returns it by the pve-qm-ipconfig format.
// If this address pool is full, Allocate returns empty string.
func (p *Segment) Allocate() string {
	ip := p.findFreeAddress()
	if ip == nil {
		// full
		return ""
	}
	return fmt.Sprintf("gw=%s,ip=%s/%d", p.Gateway, ip, p.Mask)
}
func (p *Segment) Free(ip net.IP) {
	ipv4 := p.ip2ipv4(ip)
	if !p.isUsedAddress(ipv4) {
		// Detected a bug.
		panic(fmt.Errorf("not allocated ip: %s", ip))
	}
	delete(p.allocated, ipv4.String())
}
func (p *Segment) ip2ipv4(ip net.IP) *IPv4Address {
	return newIPv4AddressByIP(
		ip,
		&net.IPNet{
			IP:   p.StartAddress,
			Mask: net.CIDRMask(int(p.Mask), 32),
		},
	)
}

// findFreeAddress finds non-allocated address and returns it.
func (p *Segment) findFreeAddress() net.IP {
	loopStopIP := p.ipCursor
	endIP := p.ip2ipv4(p.EndAddress)
	cursor := p.ipCursor
	for {
		if !(cursor.IsNetwork() || cursor.IsBroadcast()) {
			if !p.isUsedAddress(cursor) {
				// It is usable address.
				p.ipCursor = cursor
				p.allocated[cursor.String()] = struct{}{}
				return p.ipCursor.ip
			}
		}

		// Update cursor position.
		next, err := cursor.Increase()
		if err != nil {
			panic(err)
		}
		cursor = next

		if endIP.LargerThan(cursor) {
			// Reached to end.
			// Move the cursor to start address.
			cursor = p.ip2ipv4(p.StartAddress)
		}

		if loopStopIP.Equals(cursor) {
			// All addresses are allocated.
			// We can not find the free address.
			return nil
		}
	}
	panic("unreachable")
}
func (p *Segment) isUsedAddress(address *IPv4Address) bool {
	_, ok := p.allocated[address.String()]
	return ok
}

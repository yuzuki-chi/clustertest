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
}

type segmentCursor struct {
	Segment

	ipCursor  *IPv4Address
	allocated map[string]struct{}
}

func (c *segmentCursor) Init() {
	network := c.ip2ipv4(c.StartAddress).network
	if !network.Contains(c.Gateway) {
		panic(errors.Errorf("the Gateway address is out of network: gw=%s network=%s", c.Gateway, network))
	}
	if !network.Contains(c.EndAddress) {
		panic(errors.Errorf("the EndAddress is out of network: gw=%s network=%s", c.Gateway, network))
	}

	c.ipCursor = c.ip2ipv4(c.StartAddress)
	c.allocated = map[string]struct{}{}
}

// Allocate allocates an IPv4 address and returns it by the pve-qm-ipconfig format.
// If this address pool is full, Allocate returns empty string.
func (c *segmentCursor) Allocate() string {
	ip := c.findFreeAddress()
	if ip == nil {
		// full
		return ""
	}
	return fmt.Sprintf("gw=%s,ip=%s/%d", c.Gateway, ip, c.Mask)
}
func (c *segmentCursor) Free(ip net.IP) {
	ipv4 := c.ip2ipv4(ip)
	if !c.isUsedAddress(ipv4) {
		// Detected a bug.
		panic(fmt.Errorf("not allocated ip: %s", ip))
	}
	delete(c.allocated, ipv4.String())
}
func (c *segmentCursor) ip2ipv4(ip net.IP) *IPv4Address {
	return newIPv4AddressByIP(
		ip,
		&net.IPNet{
			IP:   c.StartAddress,
			Mask: net.CIDRMask(int(c.Mask), 32),
		},
	)
}

// findFreeAddress finds non-allocated address and returns it.
func (c *segmentCursor) findFreeAddress() net.IP {
	loopStopIP := c.ipCursor
	endIP := c.ip2ipv4(c.EndAddress)
	cursor := c.ipCursor
	for {
		if !(cursor.IsNetwork() || cursor.IsBroadcast()) {
			if !c.isUsedAddress(cursor) {
				// It is usable address.
				c.ipCursor = cursor
				c.allocated[cursor.String()] = struct{}{}
				return c.ipCursor.ip
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
			cursor = c.ip2ipv4(c.StartAddress)
		}

		if loopStopIP.Equals(cursor) {
			// All addresses are allocated.
			// We can not find the free address.
			return nil
		}
	}
	panic("unreachable")
}
func (c *segmentCursor) isUsedAddress(address *IPv4Address) bool {
	_, ok := c.allocated[address.String()]
	return ok
}

package addresspool

import (
	"fmt"
	"net"
	"sync"
)

var GlobalPool = NewAddressPool("global")

type AddressPool struct {
	// Identify of the address pool for debug.
	Name string

	allocated map[string]struct{}
	m         sync.Mutex
}

func NewAddressPool(name string) *AddressPool {
	return &AddressPool{
		Name:      name,
		allocated: map[string]struct{}{},
	}
}

// Allocate allocates an IPv4 address and returns it by the pve-qm-ipconfig format.
// If this address pool is full, Allocate returns empty string.
func (p *AddressPool) Allocate(segments []Segment) string {
	p.m.Lock()
	defer p.m.Unlock()

	isUsed := func(ip net.IP) bool { return p.isUsedAddress(ip) }
	for _, s := range segments {
		ip := s.FindFreeAddress(isUsed)
		if ip != nil {
			// Found a free IP address.
			p.allocated[ip.String()] = struct{}{}
			return fmt.Sprintf("gw=%s,ip=%s/%d", s.Gateway, ip, s.Mask)
		}
	}
	return ""
}
func (p *AddressPool) Free(ip net.IP) {
	p.m.Lock()
	defer p.m.Unlock()

	if !p.isUsedAddress(ip) {
		// Detected a bug.
		panic(fmt.Errorf("not allocated ip: %s", ip))
	}
	delete(p.allocated, ip.String())
}
func (p *AddressPool) isUsedAddress(ip net.IP) bool {
	_, ok := p.allocated[ip.String()]
	return ok
}

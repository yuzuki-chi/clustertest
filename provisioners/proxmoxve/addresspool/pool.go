package addresspool

import (
	"fmt"
	"github.com/pkg/errors"
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
type AddressPoolTx struct {
	P         *AddressPool
	m         sync.Mutex
	allocated []net.IP
	committed bool
	reverted  bool
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
	s, ip, ok := p.AllocateIP(segments)
	if !ok {
		return ""
	}
	return p.ToPveIPConf(s, ip)
}
func (p *AddressPool) AllocateIP(segments []Segment) (Segment, net.IP, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	isUsed := func(ip net.IP) bool { return p.isUsedAddress(ip) }
	for _, s := range segments {
		ip := s.FindFreeAddress(isUsed)
		if ip != nil {
			// Found a free IP address.
			p.allocated[ip.String()] = struct{}{}
			return s, ip, true
		}
	}
	return Segment{}, nil, false
}
func (p *AddressPool) ToPveIPConf(s Segment, ip net.IP) string {
	return fmt.Sprintf("gw=%s,ip=%s/%d", s.Gateway, ip, s.Mask)
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

// Transaction starts an transaction to allocate IP addresses atomically.
func (p *AddressPool) Transaction(fn func(tx *AddressPoolTx) error) error {
	// NOTE: MUST NOT acquire the lock on this method.
	// Because AllocateIP() and Free() are called some times while running the fn(),

	tx := AddressPoolTx{P: p}
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

// Allocate allocates an IPv4 address and returns it by the pve-qm-ipconfig format.
// If this address pool is full, Allocate returns empty string.
func (tx *AddressPoolTx) Allocate(segments []Segment) string {
	s, ip, ok := tx.AllocateIP(segments)
	if !ok {
		return ""
	}
	tx.allocated = append(tx.allocated, ip)
	return tx.P.ToPveIPConf(s, ip)
}

func (tx *AddressPoolTx) AllocateIP(segments []Segment) (Segment, net.IP, bool) {
	tx.m.Lock()
	defer tx.m.Unlock()
	tx.mustNotFinished()

	s, ip, ok := tx.P.AllocateIP(segments)
	if !ok {
		return Segment{}, nil, false
	}
	tx.allocated = append(tx.allocated, ip)
	return s, ip, true
}

// Commit allocates all reserved resources while this transaction.
func (tx *AddressPoolTx) Commit() {
	tx.m.Lock()
	defer tx.m.Unlock()
	tx.mustNotFinished()
}

// Revert releases all reserved resources while this transaction.
func (tx *AddressPoolTx) Revert() {
	tx.m.Lock()
	defer tx.m.Unlock()
	tx.mustNotFinished()
	for _, a := range tx.allocated {
		tx.P.Free(a)
	}
}

// IsFinished returns true if already committed or reverted.
// Otherwise, it returns false.
func (tx *AddressPoolTx) IsFinished() bool {
	tx.m.Lock()
	defer tx.m.Unlock()
	return tx.committed || tx.reverted
}
func (tx *AddressPoolTx) mustNotFinished() {
	if tx.committed {
		panic(errors.Errorf("this transaction has been committed"))
	}
	if tx.reverted {
		panic(errors.Errorf("this transaction has been reverted"))
	}
}

package addresspool

import (
	"fmt"
	"net"
)

type IPv4Address struct {
	// IP address
	ip net.IP
	// Network address and mask.
	network *net.IPNet
	// String of the original CIDR notation IP address.
	// This field is optional.
	original string
}

func NewIPv4Address(cidrIP string) (*IPv4Address, error) {
	ip, network, err := net.ParseCIDR(cidrIP)
	if err != nil {
		return nil, err
	}
	if ip.To4() == nil {
		return nil, fmt.Errorf("invalid IPv4 address: %s", ip)
	}
	return &IPv4Address{
		ip:       ip,
		network:  network,
		original: cidrIP,
	}, nil
}
func newIPv4AddressByIP(ip net.IP, network *net.IPNet) *IPv4Address {
	if !network.Contains(ip) {
		err := fmt.Errorf("out of network ip address: %s is not contains %s", ip, network)
		panic(err)
	}
	return &IPv4Address{
		ip:      ip,
		network: network,
	}
}
func (a *IPv4Address) String() string {
	ipnet := net.IPNet{
		IP:   a.ip,
		Mask: a.network.Mask,
	}
	return ipnet.String()
}
func (a *IPv4Address) LargerThan(address *IPv4Address) bool {
	return a.toUint32() < address.toUint32()
}
func (a *IPv4Address) Equals(address *IPv4Address) bool {
	return a.String() == address.String()
}
func (a *IPv4Address) Increase() (*IPv4Address, error) {
	oldIPNumber := a.toUint32()
	newIPNumber := oldIPNumber + 1

	// check overflow
	if oldIPNumber >= newIPNumber {
		return nil, fmt.Errorf("out of range")
	}

	newIP := a.uint32ToIP(newIPNumber)
	if a.network.Contains(newIP) {
		return newIPv4AddressByIP(newIP, a.network), nil
	}
	return nil, fmt.Errorf("out of range")
}

// IsNetwork returns true if it represents network address.
func (a *IPv4Address) IsNetwork() bool {
	ip := a.toUint32()
	bitmask := a.hostMask()
	return ip&bitmask == 0
}

// IsBroadcast returns true if it represents broadcast address.
func (a *IPv4Address) IsBroadcast() bool {
	ip := a.toUint32()
	bitmask := a.hostMask()
	return ip&bitmask == bitmask
}
func (a *IPv4Address) toUint32() uint32 {
	ipv4 := a.ip.To4()
	if ipv4 == nil {
		panic(fmt.Errorf("not ipv4: %s", a.ip.String()))
	}

	number := uint32(0)
	for i := range ipv4 {
		number = number<<8 + uint32(ipv4[i])
	}
	return number
}
func (a *IPv4Address) networkMask() uint32 {
	return ^a.hostMask()
}
func (a *IPv4Address) hostMask() uint32 {
	size, _ := a.network.Mask.Size()
	return uint32(1)<<uint(size) - 1
}
func (*IPv4Address) uint32ToIP(n uint32) net.IP {
	return net.IPv4(
		byte(n>>24),
		byte(n>>16),
		byte(n>>8),
		byte(n),
	)
}

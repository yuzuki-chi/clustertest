package addresspool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewIPv4Address(t *testing.T) {
	t.Run("should_success_when_valid_ip", func(t *testing.T) {
		ip, err := NewIPv4Address("10.0.0.1/24")
		assert.NoError(t, err)
		assert.NotNil(t, ip)
	})

	t.Run("should_fail_when_invalid_ip", func(t *testing.T) {
		ip, err := NewIPv4Address("12345")
		assert.Error(t, err)
		assert.Nil(t, ip)

		ip, err = NewIPv4Address("10.0.0.0.0/24")
		assert.Error(t, err)
		assert.Nil(t, ip)

		ip, err = NewIPv4Address("10.0.0/24")
		assert.Error(t, err)
		assert.Nil(t, ip)

		ip, err = NewIPv4Address("10.0.0.300/24")
		assert.Error(t, err)
		assert.Nil(t, ip)

		ip, err = NewIPv4Address("10.0.0.0")
		assert.Error(t, err)
		assert.Nil(t, ip)
	})

	t.Run("should_fail_when_ipv6", func(t *testing.T) {
		ip, err := NewIPv4Address("::1/24")
		assert.Error(t, err)
		assert.Nil(t, ip)
	})

}
func TestIPv4Address_Increase(t *testing.T) {
	t.Run("should_success_when_next_ip_is_exists", func(t *testing.T) {
		ip, err := NewIPv4Address("10.0.0.1/24")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, ip)

		nextIP, err := ip.Increase()
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, nextIP)
		assert.Equal(t, "10.0.0.2/24", nextIP.String())
	})

	t.Run("should_fail_when_next_ip_is_not_exist", func(t *testing.T) {
		ip, err := NewIPv4Address("10.0.0.255/24")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, ip)

		nextIP, err := ip.Increase()
		assert.EqualError(t, err, "out of range")
		assert.Nil(t, nextIP)
	})
}

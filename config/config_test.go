package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadFromBytes(t *testing.T) {
	t.Run("should_success_emtpy_", func(t *testing.T) {
		conf, err := LoadFromBytes([]byte(""))
		assert.NoError(t, err)
		assert.Equal(t, &Config{}, conf)
	})
}

package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/yuuki0xff/clustertest/models"
	"testing"
)

func TestSpecConfig_Load(t *testing.T) {
	t.Run("should_return_an_error_when_load_unsupported_type", func(t *testing.T) {
		c := &SpecConfig{
			Type: models.SpecType("invalid-type"),
			Data: nil,
		}
		spec, err := c.Load()
		assert.Nil(t, spec)
		assert.EqualError(t, err, "unsupported spec type: invalid-type")
	})

	t.Run("should_success_when_load_empty_spec", func(t *testing.T) {
		RegisteredSpecLoaders["empty"] = func(js []byte) (models.Spec, error) {
			assert.Contains(t, [][]byte{
				[]byte("null"),
				[]byte("{}"),
			}, js)
			return &testSpec{}, nil
		}

		c := &SpecConfig{
			Type: models.SpecType("empty"),
			Data: nil,
		}
		spec, err := c.Load()
		assert.IsType(t, &testSpec{}, spec)
		assert.NoError(t, err)

		c = &SpecConfig{
			Type: models.SpecType("empty"),
			Data: map[string]string{},
		}
		spec, err = c.Load()
		assert.IsType(t, &testSpec{}, spec)
		assert.NoError(t, err)
	})
}

type testSpec struct{}

func (*testSpec) String() string {
	return "testSpec"
}
func (*testSpec) Type() models.SpecType {
	return models.SpecType("testSpec")
}

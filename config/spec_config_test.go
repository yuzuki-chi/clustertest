package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/yaml"
	"testing"
)

func TestSpecConfig_Load(t *testing.T) {
	t.Run("should_return_an_error_when_load_unsupported_type", func(t *testing.T) {
		c := &SpecConfig{}
		data := []byte(`
type: invalid-type
extra: fields
`)

		err := yaml.Unmarshal(data, c)
		assert.EqualError(t, err, "unsupported type: invalid-type")
		assert.Nil(t, c.Data)
	})

	t.Run("should_success_when_load_empty_spec", func(t *testing.T) {
		c := &SpecConfig{}
		data := []byte(`
type: empty
`)
		err := yaml.Unmarshal(data, c)
		assert.NoError(t, err)
		assert.IsType(t, &emptySpec{}, c.Data)
	})

	t.Run("should_success_when_load_non_empty_spec", func(t *testing.T) {
		c := &SpecConfig{}
		data := []byte(`
type: test
cluster_address: cluster.local
token: invalid-token
nodes: 10
`)
		err := yaml.Unmarshal(data, c)
		assert.NoError(t, err)
		assert.IsType(t, &testSpec{}, c.Data)
		assert.Equal(t, &testSpec{
			ClusterAddress: "cluster.local",
			Token:          "invalid-token",
			Nodes:          10,
		}, c.Data)
	})
}

func init() {
	SpecInitializers[models.SpecType("empty")] = func() models.Spec { return &emptySpec{} }
	SpecInitializers[models.SpecType("test")] = func() models.Spec { return &testSpec{} }
}

type emptySpec struct{}

func (*emptySpec) String() string {
	return "empty"
}
func (*emptySpec) Type() models.SpecType {
	return models.SpecType("empty")
}

type testSpec struct {
	ClusterAddress string `yaml:"cluster_address"`
	Token          string
	Nodes          int
}

func (*testSpec) String() string {
	return "test"
}
func (*testSpec) Type() models.SpecType {
	return models.SpecType("test")
}

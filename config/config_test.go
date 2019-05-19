package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/yuuki0xff/clustertest/models"
	"testing"
)

func TestLoadFromBytes(t *testing.T) {
	t.Run("should_fail_when_load_empty_config", func(t *testing.T) {
		conf, err := LoadFromBytes([]byte(""))
		assert.EqualError(t, err, "unsupported config version: 0")
		assert.Nil(t, conf)
	})

	t.Run("should_fail_when_missing_name_field", func(t *testing.T) {
		conf, err := LoadFromBytes([]byte(`
version: 1
spec:
`))
		assert.EqualError(t, err, "the Config.Name is empty")
		assert.Nil(t, conf)
	})

	t.Run("should_fail_when_missing_spec_field", func(t *testing.T) {
		conf, err := LoadFromBytes([]byte(`
version: 1
name: test_config
`))
		assert.EqualError(t, err, "the Config.Spec is empty")
		assert.Nil(t, conf)
	})

	t.Run("should_success_when_loading a_valid_config", func(t *testing.T) {
		RegisteredSpecLoaders[models.SpecType("fake_spec")] = fakeSpecLoader
		RegisteredScriptLoaders[models.ScriptType("fake_script")] = fakeScriptLoader
		conf, err := LoadFromBytes([]byte(`
version: 1
name: test_config
spec:
  type: fake_spec
  fake_field1: foo
  fake_field2: bar
  nested:
    nested_field1: baz
scripts:
  main:
    type: fake_script
    fake_field1: a
    fake_field2: b
    nested:
      nested_field1: c
`))
		assert.NoError(t, err)
		assert.NotNil(t, conf)
		assert.Equal(t, "foo", conf.Spec().(*fakeSpec).FakeField1)
		assert.Equal(t, "bar", conf.Spec().(*fakeSpec).FakeField2)
		assert.Equal(t, "baz", conf.Spec().(*fakeSpec).Nested.NestedField1)
		assert.Equal(t, "a", conf.Script().(*fakeScript).FakeField1)
		assert.Equal(t, "b", conf.Script().(*fakeScript).FakeField2)
		assert.Equal(t, "c", conf.Script().(*fakeScript).Nested.NestedField1)
	})
}

func fakeSpecLoader(js []byte) (models.Spec, error) {
	s := &fakeSpec{}
	err := json.Unmarshal(js, s)
	return s, err
}

type fakeSpec struct {
	FakeField1 string `json:"fake_field1"`
	FakeField2 string `json:"fake_field2"`
	Nested     struct {
		NestedField1 string `json:"nested_field1"`
	}
}

func (s *fakeSpec) String() string {
	return "fakeSpec"
}
func (*fakeSpec) Type() models.SpecType {
	return models.SpecType("fake_spec")
}

func fakeScriptLoader(js []byte) (models.Script, error) {
	s := &fakeScript{}
	err := json.Unmarshal(js, s)
	return s, err
}

type fakeScript struct {
	FakeField1 string `json:"fake_field1"`
	FakeField2 string `json:"fake_field2"`
	Nested     struct {
		NestedField1 string `json:"nested_field1"`
	}
}

func (s *fakeScript) String() string {
	return "fakeScript"
}
func (*fakeScript) Type() models.ScriptType {
	return models.ScriptType("fake_script")
}

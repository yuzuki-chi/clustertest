package config

import (
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
		assert.EqualError(t, err, "the Config.Specs is empty")
		assert.Nil(t, conf)
	})

	t.Run("should_success_when_loading a_valid_config", func(t *testing.T) {
		conf, err := LoadFromBytes([]byte(`
version: 1
name: test_config
specs:
- type: fake_spec
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
		if !assert.NoError(t, err) {
			return
		}
		if !assert.NotNil(t, conf) {
			return
		}
		specs := conf.Specs()
		if !assert.Len(t, specs, 1) {
			return
		}

		if !assert.IsType(t, &fakeSpec{}, specs[0]) {
			return
		}
		f := specs[0].(*fakeSpec)
		assert.Equal(t, "foo", f.FakeField1)
		assert.Equal(t, "bar", f.FakeField2)
		assert.Equal(t, "baz", f.Nested.NestedField1)

		if !assert.IsType(t, &fakeScript{}, f.Scripts.Main.Data) {
			return
		}
		s := f.Scripts.Main.Data.(*fakeScript)
		assert.Equal(t, "a", s.FakeField1)
		assert.Equal(t, "b", s.FakeField2)
		assert.Equal(t, "c", s.Nested.NestedField1)
	})
}

func init() {
	SpecInitializers[models.SpecType("fake_spec")] = func() models.Spec { return &fakeSpec{} }
	ScriptInitializers[models.ScriptType("fake_script")] = func() models.Script { return &fakeScript{} }
}

type fakeSpec struct {
	FakeField1 string `yaml:"fake_field1"`
	FakeField2 string `yaml:"fake_field2"`
	Nested     struct {
		NestedField1 string `yaml:"nested_field1"`
	}
	Scripts *ScriptConfigSet
}

func (s *fakeSpec) String() string {
	return "fakeSpec"
}
func (*fakeSpec) Type() models.SpecType {
	return models.SpecType("fake_spec")
}

type fakeScript struct {
	FakeField1 string `yaml:"fake_field1"`
	FakeField2 string `yaml:"fake_field2"`
	Nested     struct {
		NestedField1 string `yaml:"nested_field1"`
	}
}

func (s *fakeScript) String() string {
	return "fakeScript"
}
func (*fakeScript) Type() models.ScriptType {
	return models.ScriptType("fake_script")
}
func (*fakeScript) SetAttr(key, value interface{}) {
	panic("not implemented")
}
func (*fakeScript) GetAttr(key interface{}) interface{} {
	panic("not implemented")
}

package config

import (
	"fmt"
	"github.com/yuuki0xff/clustertest/cmdutils"
	"github.com/yuuki0xff/clustertest/models"
)

var ScriptInitializers = map[models.ScriptType]ScriptInitializer{}

type ScriptInitializer func() models.Script

type ScriptConfigSet struct {
	Before *ScriptConfig
	Main   *ScriptConfig
	After  *ScriptConfig
}

type ScriptConfig struct {
	Type models.ScriptType
	Data models.Script
}

func (c *ScriptConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return cmdutils.UnmarshalYAMLInterface(unmarshal, func(typeName string) (concrete interface{}, err error) {
		t := models.ScriptType(typeName)
		fn, ok := ScriptInitializers[t]
		if ok {
			c.Type = t
			c.Data = fn()
			return c.Data, nil
		}
		return nil, fmt.Errorf("unsupported type: %s", typeName)
	})
}
func (c *ScriptConfig) SetAttrs(attrs map[interface{}]interface{}) *ScriptConfig {
	s := c.Get()
	if s != nil {
		for key, value := range attrs {
			s.SetAttr(key, value)
		}
	}
	return c
}
func (c *ScriptConfig) Get() models.Script {
	if c == nil {
		return nil
	}
	return c.Data
}

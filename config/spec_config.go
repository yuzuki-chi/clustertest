package config

import (
	"fmt"
	"github.com/yuuki0xff/clustertest/cmdutils"
	"github.com/yuuki0xff/clustertest/models"
)

var SpecInitializers = map[models.SpecType]SpecInitializer{}

type SpecInitializer func() models.Spec

type SpecConfig struct {
	Type models.SpecType
	Data models.Spec
}

func (c *SpecConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return cmdutils.UnmarshalYAMLInterface(unmarshal, func(typeName string) (concrete interface{}, err error) {
		t := models.SpecType(typeName)
		fn, ok := SpecInitializers[t]
		if ok {
			c.Type = t
			c.Data = fn()
			return c.Data, nil
		}
		return nil, fmt.Errorf("unsupported SpecType: %s", typeName)
	})
}

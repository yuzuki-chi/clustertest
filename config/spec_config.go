package config

import (
	"encoding/json"
	"fmt"
	"github.com/yuuki0xff/clustertest/models"
)

var RegisteredSpecLoaders = map[models.SpecType]SpecLoader{}

type SpecLoader func(json []byte) (models.Spec, error)

type SpecConfig struct {
	Type models.SpecType
	Data interface{}
}

func (c *SpecConfig) Load() (models.Spec, error) {
	fn := RegisteredSpecLoaders[c.Type]
	if fn == nil {
		return nil, fmt.Errorf("unsupported spec type: %s", c.Type)
	}
	b, err := json.Marshal(c.Data)
	if err != nil {
		panic(err)
	}
	return fn(b)
}
func (c *SpecConfig) Validate() bool {
	_, err := c.Load()
	return err == nil
}

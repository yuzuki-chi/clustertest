package config

import (
	"encoding/json"
	"fmt"
	"github.com/yuuki0xff/clustertest/models"
)

var RegisteredScriptLoaders = map[models.ScriptType]ScriptLoader{}

type ScriptLoader func(json []byte) (models.Script, error)

type ScriptConfig struct {
	Type models.ScriptType
	Data interface{}
}

func (c *ScriptConfig) Load() (models.Script, error) {
	fn := RegisteredScriptLoaders[c.Type]
	if fn == nil {
		return nil, fmt.Errorf("unsupported script type: %s", c.Type)
	}
	b, err := json.Marshal(c.Data)
	if err != nil {
		panic(err)
	}
	return fn(b)
}
func (c *ScriptConfig) MustLoad() models.Script {
	s, err := c.Load()
	if err != nil {
		panic(err)
	}
	return s
}
func (c *ScriptConfig) Validate() bool {
	_, err := c.Load()
	return err == nil
}

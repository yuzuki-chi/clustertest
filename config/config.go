package config

import (
	"fmt"
	"github.com/yuuki0xff/clustertest/models"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Version int
	Name    string
	Spec_   *SpecConfig
	Scripts struct {
		Before *ScriptConfig
		Main   *ScriptConfig
		After  *ScriptConfig
	}

	validatedSpec   models.Spec
	validatedBefore models.Script
	validatedScript models.Script
	validatedAfter  models.Script
}

func (c *Config) String() string {
	return fmt.Sprintf("Config(name=%s)", c.Name)
}
func (c *Config) Spec() models.Spec {
	return c.validatedSpec
}
func (c *Config) Before() models.Script {
	return c.validatedBefore
}
func (c *Config) Script() models.Script {
	return c.validatedScript
}
func (c *Config) After() models.Script {
	return c.validatedAfter
}

func LoadFromBytes(b []byte) (*Config, error) {
	conf := &Config{}
	err := yaml.Unmarshal(b, conf)
	if err != nil {
		return nil, err
	}

	// TODO: load spec
	// Load scripts.
	func() {
		defer func() {
			err = recover().(error)
		}()
		conf.validatedBefore = conf.Scripts.Before.MustLoad()
		conf.validatedScript = conf.Scripts.Main.MustLoad()
		conf.validatedAfter = conf.Scripts.After.MustLoad()
	}()
	if err != nil {
		return nil, err
	}
	return conf, nil
}

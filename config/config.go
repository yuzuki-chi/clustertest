package config

import (
	"errors"
	"fmt"
	"github.com/yuuki0xff/clustertest/models"
	"github.com/yuuki0xff/yaml"
)

type Config struct {
	Version int
	Name    string
	Spec_   *SpecConfig `yaml:"spec"`
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
func (c *Config) init() error {
	var err error

	if c.Version != 1 {
		return fmt.Errorf("unsupported config version: %d", c.Version)
	}
	if c.Name == "" {
		return errors.New("the Config.Name is empty")
	}
	if c.Spec_ == nil {
		return errors.New("the Config.Spec is empty")
	}
	c.validatedSpec, err = c.Spec_.Load()
	if err != nil {
		return err
	}

	// Load scripts.
	err = func() (err error) {
		defer func() {
			if obj := recover(); obj != nil {
				err = obj.(error)
			}
		}()

		if c.Scripts.Before != nil {
			c.validatedBefore = c.Scripts.Before.MustLoad()
		}
		if c.Scripts.Main == nil {
			err = errors.New("the Config.Scripts.Main is empty")
			return
		}
		c.validatedScript = c.Scripts.Main.MustLoad()
		if c.Scripts.After != nil {
			c.validatedAfter = c.Scripts.After.MustLoad()
		}
		return
	}()
	return err
}

func LoadFromBytes(b []byte) (*Config, error) {
	conf := &Config{}
	err := yaml.Unmarshal(b, conf)
	if err != nil {
		return nil, err
	}

	err = conf.init()
	if err != nil {
		return nil, err
	}
	return conf, nil
}

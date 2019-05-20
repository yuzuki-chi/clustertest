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
	Specs_  []*SpecConfig `yaml:"specs"`
}

func (c *Config) String() string {
	return fmt.Sprintf("Config(name=%s)", c.Name)
}
func (c *Config) Specs() []models.Spec {
	var specs []models.Spec
	for _, c := range c.Specs_ {
		specs = append(specs, c.Data)
	}
	return specs
}
func (c *Config) init() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version: %d", c.Version)
	}
	if c.Name == "" {
		return errors.New("the Config.Name is empty")
	}
	if len(c.Specs_) == 0 {
		return errors.New("the Config.Specs is empty")
	}
	return nil
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

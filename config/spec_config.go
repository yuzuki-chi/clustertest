package config

import "github.com/yuuki0xff/clustertest/models"

type SpecConfig struct {
	Type models.SpecType
	Data interface{}
}

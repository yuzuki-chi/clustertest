package config

import (
	"github.com/yuuki0xff/clustertest/models"
)

type ScriptConfig struct {
	Type models.ScriptType
	Data interface{}
}

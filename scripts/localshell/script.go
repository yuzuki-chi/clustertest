package localshell

import (
	"fmt"
	"github.com/yuuki0xff/clustertest/config"
	"github.com/yuuki0xff/clustertest/models"
)

const scriptType = models.ScriptType("local-shell")

func init() {
	config.ScriptInitializers[scriptType] = func() models.Script {
		return &Script{}
	}
}

type Script struct {
	Commands []string
	attr     map[interface{}]interface{}
}

func (*Script) String() string {
	return fmt.Sprintf("<%s>", scriptType)
}
func (*Script) Type() models.ScriptType {
	return scriptType
}
func (s *Script) SetAttr(key, value interface{}) {
	s.attr[key] = value
}
func (s *Script) GetAttr(key interface{}) interface{} {
	return s.attr[key]
}

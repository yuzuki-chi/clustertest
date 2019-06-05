package remoteshell

import (
	"fmt"
	"github.com/yuuki0xff/clustertest/config"
	"github.com/yuuki0xff/clustertest/models"
)

const scriptType = models.ScriptType("remote-shell")

func init() {
	config.ScriptInitializers[scriptType] = func() models.Script {
		return nil
	}
}

type Script struct {
	Commands []string
	attrs    map[interface{}]interface{}
}

func (*Script) String() string {
	return fmt.Sprintf("<%s>", scriptType)
}
func (*Script) Type() models.ScriptType {
	return scriptType
}
func (s *Script) SetAttr(key, value interface{}) {
	if s.attrs == nil {
		s.attrs = map[interface{}]interface{}{}
	}
	s.attrs[key] = value
}
func (s *Script) GetAttr(key interface{}) interface{} {
	if s.attrs == nil {
		return nil
	}
	return s.attrs[key]
}

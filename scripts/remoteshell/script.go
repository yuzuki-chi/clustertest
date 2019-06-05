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
}

func (*Script) String() string {
	return fmt.Sprintf("<%s>", scriptType)
}
func (*Script) Type() models.ScriptType {
	return scriptType
}

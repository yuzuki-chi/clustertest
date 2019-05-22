package provisioners

import "github.com/yuuki0xff/clustertest/models"

var Provisioners = map[models.SpecType]Initializer{}

type Initializer func(spec models.Spec) models.Provisioner

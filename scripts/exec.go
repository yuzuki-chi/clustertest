package scripts

import "github.com/yuuki0xff/clustertest/models"

func ExecuteBefore(p models.Provisioner, sets []*models.ScriptSet) {
	var scripts []models.Script
	for _, set := range sets {
		scripts = append(scripts, set.Before)
	}
	executeAll(p, scripts)
}

func ExecuteMain(p models.Provisioner, sets []*models.ScriptSet) {
	var scripts []models.Script
	for _, set := range sets {
		scripts = append(scripts, set.Main)
	}
	executeAll(p, scripts)
}

func ExecuteAfter(p models.Provisioner, sets []*models.ScriptSet) {
	var scripts []models.Script
	for _, set := range sets {
		scripts = append(scripts, set.After)
	}
	executeAll(p, scripts)
}

func executeAll(p models.Provisioner, scripts []models.Script) {
	for _, s := range scripts {
		e := p.ScriptExecutor(s.Type())
		result := e.Execute(s)
		// todo
		_ = result
	}
}

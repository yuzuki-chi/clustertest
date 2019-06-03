package scripts

import "github.com/yuuki0xff/clustertest/models"

func ExecuteBefore(p models.Provisioner, sets []*models.ScriptSet) models.ScriptResult {
	var scripts []models.Script
	for _, set := range sets {
		scripts = append(scripts, set.Before)
	}
	return executeAll(p, scripts)
}

func ExecuteMain(p models.Provisioner, sets []*models.ScriptSet) models.ScriptResult {
	var scripts []models.Script
	for _, set := range sets {
		scripts = append(scripts, set.Main)
	}
	return executeAll(p, scripts)
}

func ExecuteAfter(p models.Provisioner, sets []*models.ScriptSet) models.ScriptResult {
	var scripts []models.Script
	for _, set := range sets {
		scripts = append(scripts, set.After)
	}
	return executeAll(p, scripts)
}

func executeAll(p models.Provisioner, scripts []models.Script) models.ScriptResult {
	for _, s := range scripts {
		e := p.ScriptExecutor(s.Type())
		result := e.Execute(s)
		// todo
		_ = result
	}
	// todo
	panic("not impl")
}

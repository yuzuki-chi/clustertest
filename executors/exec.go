package executors

import (
	"github.com/republicprotocol/co-go"
	"github.com/yuuki0xff/clustertest/models"
	"sync"
)

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
	m := sync.Mutex{}
	mr := &MergedResult{}

	co.ParForAll(scripts, func(i int) {
		s := scripts[i]
		if s == nil {
			return
		}
		e := p.ScriptExecutor(s.Type())
		result := e.Execute(s)

		m.Lock()
		mr.Append(result)
		m.Unlock()
	})
	return mr
}

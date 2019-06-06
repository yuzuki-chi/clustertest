package main

import "github.com/yuuki0xff/clustertest/config"

func loadConfigs(args []string) ([]*config.Config, error) {
	if len(args) == 0 {
		// Load a config file from "./".
		args = []string{"./"}
	}
	return config.LoadFromDirsOrFiles(args)
}

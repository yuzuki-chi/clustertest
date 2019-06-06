package main

import "github.com/yuuki0xff/clustertest/config"

func findConfigs(args []string) ([]string, error) {
	if len(args) == 0 {
		// Load a config file from "./".
		args = []string{"./"}
	}
	return config.FindConfigFromDirsOrFiles(args)
}

func loadConfigs(args []string) ([]*config.Config, error) {
	files, err := findConfigs(args)
	if err != nil {
		return nil, err
	}
	return config.LoadFromFiles(files)
}

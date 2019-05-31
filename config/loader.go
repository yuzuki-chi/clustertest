package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// Load config files from directories/files.
func LoadFromDirsOrFiles(paths []string) ([]*Config, error) {
	var dirs []string
	var files []string

	for _, path := range paths {
		finfo, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if finfo.IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
		}
	}

	tasks1, err := LoadFromDirs(dirs)
	if err != nil {
		return nil, err
	}
	tasks2, err := LoadFromFiles(files)
	if err != nil {
		return nil, err
	}
	tasks := append(tasks1, tasks2...)
	return tasks, nil
}

// Load config files from directories.
func LoadFromDirs(dirs []string) ([]*Config, error) {
	var files []string
	for _, dir := range dirs {
		file := findByFileNmaes([]string{
			path.Join(dir, ".clustertest.yml"),
			path.Join(dir, ".clustertest.yaml"),
			path.Join(dir, "clustertest.yml"),
			path.Join(dir, "clustertest.yaml"),
		})
		if file == "" {
			// no match
			return nil, fmt.Errorf("not found config file on %s directory", dir)
		}
		// match
		files = append(files, file)
	}
	return LoadFromFiles(files)
}

// Load config files from files.
func LoadFromFiles(files []string) ([]*Config, error) {
	var confs []*Config
	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		conf, err := LoadFromBytes(b)
		if err != nil {
			return nil, err
		}
		confs = append(confs, conf)
	}
	return confs, nil
}

// isExist returns true if specified filename is exist.
func isExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// findByFileNmaes finds a file that match the file path patterns and returns the file path of the first found..
// If no matching files ware found, returns the empty string.
func findByFileNmaes(files []string) string {
	for _, file := range files {
		if isExist(file) {
			return file
		}
	}
	return ""
}

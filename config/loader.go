package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// Find config file paths from directories/files.
func FindConfigFromDirsOrFiles(paths []string) ([]string, error) {
	var dirs []string
	var files []string

	for _, p := range paths {
		finfo, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if finfo.IsDir() {
			dirs = append(dirs, p)
		} else {
			files = append(files, p)
		}
	}

	for _, dir := range dirs {
		f := findByFileNmaes([]string{
			path.Join(dir, ".clustertest.yml"),
			path.Join(dir, ".clustertest.yaml"),
			path.Join(dir, "clustertest.yml"),
			path.Join(dir, "clustertest.yaml"),
		})
		if f == "" {
			// no match
			return nil, fmt.Errorf("not found config file on %s directory", dir)
		}
		files = append(files, f)
	}

	return files, nil
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

package main

import (
	"github.com/yuuki0xff/clustertest/models"
	"io/ioutil"
)

type FileTask struct {
	name string
	data []byte
}

func newTaskFromFile(name string) (models.Task, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	return &FileTask{name, data}, nil
}
func (t *FileTask) String() string {
	return t.name
}
func (t *FileTask) SpecData() []byte {
	return t.data
}

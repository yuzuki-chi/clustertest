package cmdutils

import (
	"encoding/json"
	"errors"
)

// UnmarshalJSONInterface unmarshal the interface.
//
//   Input: {"type": "foo",  "extra-keys": "..."}
//   Output: v = ConcreteType{...}
//
// TODO: improve docs
func UnmarshalJSONInterface(js []byte, fn func(typeName string) (concrete interface{}, err error)) error {
	var m map[string]interface{}
	err := json.Unmarshal(js, &m)
	if err != nil {
		return err
	}

	typeName := m["type"]
	if typeName == "" {
		return errors.New(`missing "type" field`)
	}
	if t, ok := typeName.(string); !ok {
		return errors.New(`"type" field is not string`)
	} else {
		concrete, err := fn(t)
		if err != nil {
			return err
		}
		return json.Unmarshal(js, concrete)
	}
}

func UnmarshalYAMLInterface(unmarshal func(interface{}) error, fn func(typeName string) (concrete interface{}, err error)) error {
	var m map[string]interface{}
	err := unmarshal(&m)
	if err != nil {
		return err
	}

	typeName := m["type"]
	if typeName == "" {
		return errors.New(`missing "type" field`)
	}
	if t, ok := typeName.(string); !ok {
		return errors.New(`"type" field is not string`)
	} else {
		concrete, err := fn(t)
		if err != nil {
			return err
		}
		return unmarshal(concrete)
	}
}

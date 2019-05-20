package cmdutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type Foo struct {
	Bar AbstractUnmarshaller
}

type Abstract interface{}
type Concrete1 struct{ C1 string }
type Concrete2 struct{ C2 string }

type AbstractUnmarshaller struct {
	Data Abstract
}

func (a *AbstractUnmarshaller) UnmarshalJSON(js []byte) error {
	return UnmarshalJSONInterface(js, func(typeName string) (interface{}, error) {
		switch typeName {
		case "concrete1":
			a.Data = &Concrete1{}
		case "concrete2":
			a.Data = &Concrete2{}
		default:
			return nil, errors.New("invalid type")
		}
		return a.Data, nil
	})
}

func ExampleUnmarshalJSONInterface() {
	var foo Foo

	js := []byte(`{"bar": {"type": "concrete1", "c1": "ok"}}`)
	err := json.Unmarshal(js, &foo)
	if err != nil {
		panic(err)
	}

	fmt.Println(reflect.TypeOf(foo.Bar.Data))
	fmt.Printf("%+v\n", foo.Bar.Data)
	// Output:
	// *cmdutils.Concrete1
	// &{C1:ok}
}

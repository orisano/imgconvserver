package imgconvserver

import "reflect"

var TypeMap map[string]reflect.Value

func init() {
	TypeMap = make(map[string]reflect.Value)

	TypeMap["int"] = reflect.ValueOf(int(0))
	TypeMap["string"] = reflect.ValueOf("")
}

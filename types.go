package pipeline

import (
	"fmt"
	"reflect"
)

//TargetOf returns the type pointed to by ptr, or otherwise panics
func TargetOf(ptr interface{}) reflect.Type {
	typ := reflect.TypeOf(ptr)
	if typ.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("expected pointer type, got `%s`", typ))
	}

	return typ.Elem()
}

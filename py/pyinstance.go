package py

/*
#cgo pkg-config: python-2.7
#include "Python.h"
*/
import "C"
import (
	"pfi/sensorbee/sensorbee/data"
)

// ObjectInstance is a bind of binding Python instance, used as `PyInstance`.
type ObjectInstance struct {
	Object
}

// Call calls `name` function.
//  argument type: ...data.Value
//  return type:   data.Value
func (ins *ObjectInstance) Call(name string, args ...data.Value) (data.Value, error) {
	return invoke(ins.p, name, args)
}

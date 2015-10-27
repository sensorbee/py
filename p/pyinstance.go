package p

/*
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
	return invoke(ins.p, name, args...)
}

// CallDirect calls `name` function.
//  argument type: ...data.Value
//  return type:   Object.
//
// This method is suitable for getting the instance object that called method returned.
func (ins *ObjectInstance) CallDirect(name string, args ...data.Value) (Object, error) {
	return invokeDirect(ins.p, name, args...)
}

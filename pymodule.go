package py

/*
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"pfi/sensorbee/sensorbee/data"
	"runtime"
	"unsafe"
)

// ObjectModule is a bind of `PyObject`, used as `PyModule`
type ObjectModule struct {
	Object
}

// LoadModule loads `name` module. The module needs to be placed at `sys.path`.
// User can set optional `sys.path` using `ImportSysAndAppendPath`
func LoadModule(name string) (ObjectModule, error) {
	cModule := C.CString(name)
	defer C.free(unsafe.Pointer(cModule))

	type Result struct {
		val ObjectModule
		err error
	}
	ch := make(chan *Result, 1)
	go func() {
		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)

		pyMdl := C.PyImport_ImportModule(cModule)
		if pyMdl == nil {
			ch <- &Result{ObjectModule{}, fmt.Errorf(
				"fail to load '%v' module: %v", name, getPyErr())}
			return
		}

		ch <- &Result{ObjectModule{Object{p: pyMdl}}, nil}
	}()
	res := <-ch

	return res.val, res.err
}

// NewInstance returns `name` constructor.
//
//  ```python
//  class Sample(object):
//  	def __init__(self, a, b=5):
//  		# initializing
//  ```
//
// To get that "Sample" python instance, callers use this function like:
//  NewInstance("Sample", data.Value, data.Int)
// or
//  NewInstance("Sample", data.Value) // value "b" is optional and will set 5
func (m *ObjectModule) NewInstance(name string, args ...data.Value) (
	ObjectInstance, error) {
	return newInstance(m, name, nil, args...)
}

// NewInstanceWithKwd returns 'name' constructor with named arguments.
//
//  ```python
//  class Sample(object):
//  	def __init__(self, a, b=5, **c):
//  		# initializing
//  ```
//
// To get that "Sample" python instance, callers use a map object as named
// arguments, like:
//
//  arg1 := data.Map{
//  	"a":     data.Value, // ex) data.String("v1")
//  	"b":     data.Int,   // ex) data.Int(5)
//  	"hoge1": data.Value, // ex) data.Float(100.0)
//  	"hoge2": data.Value, // ex) data.True
//  }
// `arg1` is same as `Sawmple(a-'v1', b=5, hoge1=100.0, hoge2=True)`.
//
//  arg2 := data.Map{
//  	"a":     data.Value, // ex) data.String("v1")
//  	"hoge1": data.Value, // ex) data.Float(100.0)
//  	"hoge2": data.Value, // ex) data.True
//  }
// `arg2` is same as `Sample(a='v1', hoge1=100.0, hoge2=True)`, and `self.b`
// will be set default value (=5).
//
//  arg3 := data.Map{
//  	"a": data.Value, // ex) data.String("v1")
//  }
// `arg3` is same as `Sample(a='v1')`, `self.b` will be set default value (=5),
// and `self.c` will be set `{}`
func (m *ObjectModule) NewInstanceWithKwd(name string, kwdArgs data.Map) (
	ObjectInstance, error) {
	return newInstance(m, name, kwdArgs)
}

// GetClass returns `name` class instance.
// User needs to call DecRef when finished using instance.
func (m *ObjectModule) GetClass(name string) (ObjectInstance, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	type Result struct {
		val ObjectInstance
		err error
	}
	ch := make(chan *Result, 1)
	go func() {
		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)

		pyInstance := C.PyObject_GetAttrString(m.p, cName)
		if pyInstance == nil {
			ch <- &Result{ObjectInstance{}, fmt.Errorf(
				"fail to get '%v' instance: %v", name, getPyErr())}
			return
		}
		ch <- &Result{ObjectInstance{Object{p: pyInstance}}, nil}
	}()
	res := <-ch

	return res.val, res.err
}

// Call calls `name` function.
//  argument type: ...data.Value
//  return type:   data.Value
func (m *ObjectModule) Call(name string, args ...data.Value) (data.Value, error) {
	return invoke(m.p, name, nil, args...)
}

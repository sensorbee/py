package py

/*
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"pfi/sensorbee/py/mainthread"
	"pfi/sensorbee/sensorbee/data"
	"unsafe"
)

// ObjectModule is a bind of `PyObject`, used as `PyModule`
type ObjectModule struct {
	Object
}

// LoadModule loads `name` module. The module needs to be placed at `sys.path`.
// User can set optional `sys.path` using `mainthread.AppendSysPath`
func LoadModule(name string) (ObjectModule, error) {
	cModule := C.CString(name)
	defer C.free(unsafe.Pointer(cModule))

	type Result struct {
		val ObjectModule
		err error
	}
	ch := make(chan *Result)
	mainthread.Exec(func() {
		pyMdl := C.PyImport_ImportModule(cModule)
		if pyMdl == nil {
			ch <- &Result{ObjectModule{}, fmt.Errorf(
				"fail to load '%v' module: %v", name, getPyErr())}
			return
		}

		ch <- &Result{ObjectModule{Object{p: pyMdl}}, nil}
	})
	res := <-ch

	return res.val, res.err
}

// NewInstance returns 'name' constructor with named arguments.
//
//  class Sample(object):
//  	def __init__(self, a, b=5, **c):
//  		# initializing
//
// To get that "Sample" python instance, callers use a map object as named
// arguments, and set `kwdArgs`, like:
//
//  arg1 := data.Map{
//  	"a":     data.Value, // ex) data.String("v1")
//  	"b":     data.Int,   // ex) data.Int(5)
//  	"hoge1": data.Value, // ex) data.Float(100.0)
//  	"hoge2": data.Value, // ex) data.True
//  }
//
// `arg1` is same as `Sample(a-'v1', b=5, hoge1=100.0, hoge2=True)`.
//
//  arg2 := data.Map{
//  	"a":     data.Value, // ex) data.String("v1")
//  	"hoge1": data.Value, // ex) data.Float(100.0)
//  	"hoge2": data.Value, // ex) data.True
//  }
//
// `arg2` is same as `Sample(a='v1', hoge1=100.0, hoge2=True)`, and `self.b`
// will be set default value (=5).
//
//  arg3 := data.Map{
//  	"a": data.Value, // ex) data.String("v1")
//  }
//
// `arg3` is same as `Sample(a='v1')`, `self.b` will be set default value (=5),
// and `self.c` will be set `{}`
func (m *ObjectModule) NewInstance(name string, args []data.Value, kwdArgs data.Map) (
	ObjectInstance, error) {
	type Result struct {
		val ObjectInstance
		err error
	}
	ch := make(chan *Result)
	mainthread.Exec(func() {
		r, err := newInstance(m, name, args, kwdArgs)
		ch <- &Result{r, err}
	})
	res := <-ch
	return res.val, res.err
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
	ch := make(chan *Result)
	mainthread.Exec(func() {
		pyInstance := C.PyObject_GetAttrString(m.p, cName)
		if pyInstance == nil {
			ch <- &Result{ObjectInstance{}, fmt.Errorf(
				"fail to get '%v' instance: %v", name, getPyErr())}
			return
		}
		ch <- &Result{ObjectInstance{Object{p: pyInstance}}, nil}
	})
	res := <-ch

	return res.val, res.err
}

// Call calls `name` function. This function is supported for module method of
// python.
func (m *ObjectModule) Call(name string, args ...data.Value) (data.Value, error) {
	type Result struct {
		val data.Value
		err error
	}
	ch := make(chan *Result)
	mainthread.Exec(func() {
		v, err := invoke(m.p, name, args, nil)
		ch <- &Result{v, err}
	})
	res := <-ch
	return res.val, res.err
}

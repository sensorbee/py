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

// ObjectInstance is a bind of Python instance, used as `PyInstance`.
type ObjectInstance struct {
	Object
}

// Call calls `name` function.
//  argument type: ...data.Value
//  return type:   data.Value
func (ins *ObjectInstance) Call(name string, args ...data.Value) (data.Value,
	error) {
	return invoke(ins.p, name, nil, args...)
}

// CallDirect calls `name` function.
//  argument type: ...data.Value
//  return type:   Object.
//
// This method is suitable for getting the instance object that called method
// returned.
func (ins *ObjectInstance) CallDirect(name string, args ...data.Value) (Object,
	error) {
	return invokeDirect(ins.p, name, nil, args...)
}

// CallDirectWithKwd calls `name` function with named arguments and returns the
// function's return value.
func (ins *ObjectInstance) CallDirectWithKwd(name string, kwdArgs data.Map) (Object,
	error) {
	return invokeDirect(ins.p, name, kwdArgs)
}

func newInstance(m *ObjectModule, name string, kwdArgs data.Map,
	args ...data.Value) (ObjectInstance, error) {

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
				"fail to get '%v' class: %v", name, getPyErr())}
			return
		}
		defer C.Py_DecRef(pyInstance)

		// named arguments
		var pyKwdArg *C.PyObject
		if kwdArgs == nil || len(kwdArgs) == 0 {
			pyKwdArg = nil
		} else {
			o, err := newPyObj(kwdArgs)
			if err != nil {
				ch <- &Result{ObjectInstance{}, fmt.Errorf(
					"fail to convert named arguments in creating '%v' instance: %v",
					name, err.Error())}
				return
			}
			pyKwdArg = o.p
		}

		// no named arguments
		pyArg := C.PyTuple_New(C.Py_ssize_t(len(args)))
		if pyArg == nil {
			ch <- &Result{ObjectInstance{}, getPyErr()}
			return
		}
		defer C.Py_DecRef(pyArg)

		for i, v := range args {
			o, err := newPyObj(v)
			if err != nil {
				ch <- &Result{ObjectInstance{}, fmt.Errorf(
					"fail to convert non named arguments in creating '%v' instance: %v",
					name, err.Error())}
				return
			}
			C.PyTuple_SetItem(pyArg, C.Py_ssize_t(i), o.p)
		}

		ret := C.PyObject_Call(pyInstance, pyArg, pyKwdArg)
		if ret == nil {
			ch <- &Result{ObjectInstance{}, fmt.Errorf(
				"fail to create '%v' instance: %v", name, getPyErr())}
			return
		}
		ch <- &Result{ObjectInstance{Object{p: ret}}, nil}
	}()
	res := <-ch

	return res.val, res.err
}

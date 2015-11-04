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
	if ins.p == nil {
		return nil, fmt.Errorf("ins.p of %p is nil while calling %s", ins, name)
	}
	return invoke(ins.p, name, args, nil)
}

// CallDirect calls `name` function.
//  argument type: ...data.Value
//  return type:   Object.
//
// This method is suitable for getting the instance object that called method
// returned.
func (ins *ObjectInstance) CallDirect(name string, args []data.Value,
	kwdArg data.Map) (Object, error) {
	return invokeDirect(ins.p, name, args, kwdArg)
}

func newInstance(m *ObjectModule, name string, args []data.Value, kwdArgs data.Map) (
	ObjectInstance, error) {

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

		// no named arguments
		pyArg, err := convertArgsGo2Py(args)
		if err != nil {
			ch <- &Result{ObjectInstance{}, fmt.Errorf(
				"fail to convert non named arguments in creating '%v' instance: %v",
				name, err.Error())}
			return
		}
		defer pyArg.decRef()

		// named arguments
		var pyKwdArg *C.PyObject
		if len(kwdArgs) == 0 {
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

		ret := C.PyObject_Call(pyInstance, pyArg.p, pyKwdArg)
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

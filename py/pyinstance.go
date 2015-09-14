package py

/*
#cgo pkg-config: python-2.7
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"pfi/sensorbee/sensorbee/data"
	"runtime"
)

// ObjectInstance is a bind of binding Python instance, used as `PyInstance`.
type ObjectInstance struct {
	Object
}

// Call calls `name` function.
//  argument type: ...data.Value
//  return type:   data.Value
func (ins *ObjectInstance) Call(name string, args ...data.Value) (data.Value, error) {
	type Result struct {
		val data.Value
		err error
	}
	ch := make(chan *Result, 1)
	go func() {
		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)

		var res data.Value
		pyFunc, err := getPyInstanceFunc(ins, name)
		if err != nil {
			ch <- &Result{res, fmt.Errorf("%v at '%v'", err.Error(), name)}
			return
		}
		defer pyFunc.decRef()

		pyArg := C.PyTuple_New(C.Py_ssize_t(len(args)))
		defer C.Py_DecRef(pyArg)

		for i, v := range args {
			o := newPyObj(v)
			C.PyTuple_SetItem(pyArg, C.Py_ssize_t(i), o.p)
		}
		// TODO: defer o.decRef()

		ret, err := pyFunc.CallObject(Object{p: pyArg})
		if ret.p == nil && err != nil {
			ch <- &Result{res, fmt.Errorf("%v in '%v'", err.Error(), name)}
			return
		}
		defer ret.decRef()

		ch <- &Result{fromPyTypeObject(ret.p), nil}
	}()
	res := <-ch

	return res.val, res.err
}

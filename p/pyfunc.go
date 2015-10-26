package py

/*
#include "Python.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"pfi/sensorbee/sensorbee/data"
	"runtime"
	"unsafe"
)

// ObjectFunc is a bind of `PyObject` used as `PyFunc`
type ObjectFunc struct {
	Object
}

// invokeDirect calls name's function. User needs to call DecRef.
// This returns an Object even if result values are more than one.
// For example, use to get the object of the class instance that method returned.
func invokeDirect(pyObj *C.PyObject, name string, args ...data.Value) (Object, error) {
	type Result struct {
		val Object
		err error
	}
	ch := make(chan *Result, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- &Result{Object{}, fmt.Errorf("cannot call '%v' due to panic: %v", name, r)}
			}
		}()

		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)

		var res Object
		pyFunc, err := getPyFunc(pyObj, name)
		if err != nil {
			ch <- &Result{res, fmt.Errorf("%v at '%v'", err.Error(), name)}
			return
		}
		defer pyFunc.decRef()

		pyArg := C.PyTuple_New(C.Py_ssize_t(len(args)))
		defer C.Py_DecRef(pyArg)

		for i, v := range args {
			o, err := newPyObj(v)
			if err != nil {
				ch <- &Result{res, fmt.Errorf("%v at '%v'", err.Error(), name)}
				return
			}
			// PyTuple object takes over the value's reference, and not need to
			// decrease reference counter.
			C.PyTuple_SetItem(pyArg, C.Py_ssize_t(i), o.p)
		}

		ret, err := pyFunc.callObject(Object{p: pyArg})
		if err != nil {
			ch <- &Result{res, fmt.Errorf("%v in '%v'", err.Error(), name)}
			return
		}

		ch <- &Result{ret, err}
	}()
	res := <-ch

	return res.val, res.err
}

// invoke name's function. TODO should be placed at internal package.
func invoke(pyObj *C.PyObject, name string, args ...data.Value) (data.Value, error) {
	type Result struct {
		val data.Value
		err error
	}
	ch := make(chan *Result, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- &Result{data.Null{}, fmt.Errorf("cannot call '%v' due to panic: %v", name, r)}
			}
		}()

		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)

		var res data.Value
		pyFunc, err := getPyFunc(pyObj, name)
		if err != nil {
			ch <- &Result{res, fmt.Errorf("%v at '%v'", err.Error(), name)}
			return
		}
		defer pyFunc.decRef()

		pyArg, err := convertArgsGo2Py(args)
		if err != nil {
			ch <- &Result{res, fmt.Errorf("%v at '%v'", err.Error(), name)}
			return
		}
		defer pyArg.decRef()

		ret, err := pyFunc.callObject(pyArg)
		if err != nil {
			ch <- &Result{res, fmt.Errorf("%v in '%v'", err.Error(), name)}
			return
		}
		defer ret.decRef()

		po, err := fromPyTypeObject(ret.p)
		ch <- &Result{po, err}
	}()
	res := <-ch

	return res.val, res.err
}

// TODO should be placed at internal package
func getPyFunc(pyObj *C.PyObject, name string) (ObjectFunc, error) {
	cFunc := C.CString(name)
	defer C.free(unsafe.Pointer(cFunc))

	pyFunc := C.PyObject_GetAttrString(pyObj, cFunc)
	if pyFunc == nil {
		return ObjectFunc{}, errors.New("cannot load function")
	}

	if ok := C.PyCallable_Check(pyFunc); ok == 0 {
		return ObjectFunc{}, errors.New("cannot call function")
	}

	return ObjectFunc{Object{p: pyFunc}}, nil
}

// callObject executes python function, using `PyObject_CallObject`. Returns a
// `PyObject` even if result values are more than one. When a value will be set
// directory, and values will be set as a `PyTuple` object.
func (f *ObjectFunc) callObject(arg Object) (po Object, err error) {
	po = Object{
		p: C.PyObject_CallObject(f.p, arg.p),
	}
	if po.p == nil {
		err = getPyErr()
	}
	return
}

func convertArgsGo2Py(args []data.Value) (Object, error) {
	pyArg := C.PyTuple_New(C.Py_ssize_t(len(args)))
	shouldDecRef := true
	defer func() {
		if shouldDecRef {
			C.Py_DecRef(pyArg)
		}
	}()
	for i, v := range args {
		o, err := newPyObj(v)
		if err != nil {
			return Object{}, err
		}
		// PyTuple object takes over the value's reference, and not need to
		// decrease reference counter.
		C.PyTuple_SetItem(pyArg, C.Py_ssize_t(i), o.p)
	}
	shouldDecRef = false
	return Object{pyArg}, nil
}

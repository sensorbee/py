package py

/*
#cgo pkg-config: python-2.7
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

		pyMdl, err := C.PyImport_ImportModule(cModule)
		if pyMdl == nil {
			ch <- &Result{ObjectModule{}, fmt.Errorf("cannot load '%v' module: %v", name, err)}
			return
		}

		ch <- &Result{ObjectModule{Object{p: pyMdl}}, nil}
	}()
	res := <-ch

	return res.val, res.err
}

// GetInstance returns `name` constructor.
func (m *ObjectModule) NewInstance(name string, args ...data.Value) (ObjectModule, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	type Result struct {
		val ObjectModule
		err error
	}
	ch := make(chan *Result, 1)
	go func() {
		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)

		pyInstance := C.PyObject_GetAttrString(m.p, cName)
		if pyInstance == nil {
			ch <- &Result{ObjectModule{}, fmt.Errorf("cannot create '%v' instance", name)}
			return
		}
		defer C.Py_DecRef(pyInstance)

		pyArg := C.PyTuple_New(C.Py_ssize_t(len(args)))
		defer C.Py_DecRef(pyArg)

		for i, v := range args {
			o := newPyObj(v)
			C.PyTuple_SetItem(pyArg, C.Py_ssize_t(i), o.p)
		}

		// get constructor (called `__init__(self)`)
		ret := C.PyObject_CallObject(pyInstance, pyArg)
		ch <- &Result{ObjectModule{Object{p: ret}}, nil}
	}()
	res := <-ch

	return res.val, res.err
}

// Call calls `name` function.
//  argument type: ...data.Value
//  return type:   data.Value
func (m *ObjectModule) Call(name string, args ...data.Value) (data.Value, error) {
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
		pyFunc, err := m.getPyFunc(name)
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

func (m *ObjectModule) getPyFunc(name string) (ObjectFunc, error) {
	cFunc := C.CString(name)
	defer C.free(unsafe.Pointer(cFunc))

	pyFunc := C.PyObject_GetAttrString(m.p, cFunc)
	if pyFunc == nil {
		return ObjectFunc{}, errors.New("cannot load function")
	}

	if ok := C.PyCallable_Check(pyFunc); ok == 0 {
		return ObjectFunc{}, errors.New("cannot call function")
	}

	return ObjectFunc{Object{p: pyFunc}}, nil
}

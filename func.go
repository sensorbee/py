package py

/*
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"gopkg.in/sensorbee/sensorbee.v0/data"
	"unsafe"
)

// ObjectFunc is a bind of `PyObject` used as `PyFunc`
type ObjectFunc struct {
	Object

	name string
}

// invokeDirect calls name's function. User needs to call DecRef.
// This returns an Object even if there're multiple values returned from python.
// For example, use to get the object of the class instance that method returned.
func invokeDirect(pyObj *C.PyObject, name string, args []data.Value,
	kwdArgs data.Map) (resultObject Object, err error) {
	pyFunc, err := getPyFunc(pyObj, name)
	if err != nil {
		return Object{}, fmt.Errorf("fail to get '%v' function: %v", name,
			err.Error())
	}
	defer pyFunc.decRef()

	return pyFunc.call(args, kwdArgs)
}

// invoke name's function. TODO should be placed at internal package.
func invoke(pyObj *C.PyObject, name string, args []data.Value, kwdArgs data.Map) (
	data.Value, error) {
	ret, err := invokeDirect(pyObj, name, args, kwdArgs)
	if err != nil {
		return nil, err
	}
	defer ret.decRef()

	return fromPyTypeObject(ret.p)
}

func getPyFunc(pyObj *C.PyObject, name string) (ObjectFunc, error) {
	cFunc := C.CString(name)
	defer C.free(unsafe.Pointer(cFunc))

	pyFunc := C.PyObject_GetAttrString(pyObj, cFunc)
	if pyFunc == nil {
		return ObjectFunc{}, getPyErr()
	}

	if ok := C.PyCallable_Check(pyFunc); ok == 0 {
		return ObjectFunc{}, fmt.Errorf("'%v' is not callable object", name)
	}

	return ObjectFunc{
		Object: Object{p: pyFunc},
		name:   name,
	}, nil
}

// TODO: provide Call which acquires GIL

func (f *ObjectFunc) call(args []data.Value, kwdArgs data.Map) (result Object, resErr error) {
	defer func() {
		if r := recover(); r != nil {
			resErr = fmt.Errorf("cannot call '%v' due to panic: %v", f.name, r)
		}
	}()

	// no named arguments
	pyArg, err := convertArgsGo2Py(args)
	if err != nil {
		return Object{}, fmt.Errorf(
			"fail to convert argument in calling '%v' function: %v", f.name,
			err.Error())
	}
	defer pyArg.decRef()

	// named arguments
	var ret Object
	if len(kwdArgs) == 0 {
		ret, err = f.callObject(pyArg)
	} else {
		pyKwdArg, localErr := newPyObj(kwdArgs)
		if localErr != nil {
			return Object{}, fmt.Errorf(
				"fail to convert named arguments in calling '%v' function: %v",
				f.name, localErr)
		}
		defer pyKwdArg.decRef()
		ret, err = f.callObjectWithKwd(pyKwdArg, pyArg)
	}

	if err != nil {
		return Object{}, fmt.Errorf("fail to call '%v' function: %v", f.name,
			err.Error())
	}
	return ret, nil
}

// callObject executes python function, using `PyObject_CallObject`. Returns a
// `PyObject` even if result values are more than one. When a value will be set
// directory, and values will be set as a `PyTuple` object.
func (f *ObjectFunc) callObject(arg Object) (Object, error) {
	po := C.PyObject_CallObject(f.p, arg.p)
	if po == nil {
		return Object{}, getPyErr()
	}
	return Object{p: po}, nil
}

// callObjectWithKwd executes python function, using `PyObject_Call`. Error
// specification is same as `callObject`.
func (f *ObjectFunc) callObjectWithKwd(kwdArg Object, arg Object) (Object, error) {
	po := C.PyObject_Call(f.p, arg.p, kwdArg.p)
	if po == nil {
		return Object{}, getPyErr()
	}
	return Object{p: po}, nil
}

func convertArgsGo2Py(args []data.Value) (Object, error) {
	pyArg := C.PyTuple_New(C.Py_ssize_t(len(args)))
	if pyArg == nil {
		return Object{}, getPyErr()
	}
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

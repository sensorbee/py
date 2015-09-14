package py

/*
#cgo pkg-config: python-2.7
#include "Python.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// ObjectFunc is a bind of `PyObject` used as `PyFunc`
type ObjectFunc struct {
	Object
}

func getPyModuleFunc(mdl *ObjectModule, name string) (ObjectFunc, error) {
	return getPyFunc(mdl.p, name)
}

func getPyInstanceFunc(ins *ObjectInstance, name string) (ObjectFunc, error) {
	return getPyFunc(ins.p, name)
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

// CallObject executes python function, using `PyObject_CallObject`. Returns a
// `PyObject` even if result values are more thane one. When a value will be set
// directory, and values will be set as a `PyTuple` object.
func (f *ObjectFunc) CallObject(arg Object) (po Object, err error) {
	po = Object{}
	pyValue, err := C.PyObject_CallObject(f.p, arg.p)
	if pyValue == nil && err != nil {
		err = fmt.Errorf("call function error: %v", err)
	} else {
		po.p = pyValue
	}
	return po, err
}

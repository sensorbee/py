package py

/*
#cgo darwin pkg-config: python-2.7
#include "Python.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"pfi/sensorbee/sensorbee/data"
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

	pyMdl, err := C.PyImport_ImportModule(cModule)
	if pyMdl == nil {
		return ObjectModule{}, fmt.Errorf("cannot load '%v' module: %v", name, err)
	}
	return ObjectModule{Object{p: pyMdl}}, nil
}

// GetInstance returns `name` constructor.
func (m *ObjectModule) GetInstance(name string, args ...data.Value) (ObjectModule, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	pyInstance := C.PyObject_GetAttrString(m.p, cName)
	if pyInstance == nil {
		return ObjectModule{}, fmt.Errorf("cannot create '%v' instance", name)
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
	return ObjectModule{Object{p: ret}}, nil
}

/* TODO should be use reflection and `...interface{}` */

// CallMapString calls `name` function
func (m *ObjectModule) CallMapString(name string, ma data.Map) (
	string, error) {
	pyFunc, err := m.getPyFunc(name)
	if err != nil {
		return "", fmt.Errorf("%v at '%v'", err.Error(), name)
	}
	defer pyFunc.DecRef()

	pyArg := C.PyTuple_New(1)
	defer C.Py_DecRef(pyArg)
	pyValue := newPyObj(ma)
	defer pyValue.DecRef()
	C.PyTuple_SetItem(pyArg, 0, pyValue.p)

	ret, err := pyFunc.CallObject(Object{p: pyArg})
	if ret.p == nil && err != nil {
		return "", fmt.Errorf("%v in '%v'", err.Error(), name)
	}

	charPtr := C.PyString_AsString(ret.p)
	return C.GoString(charPtr), nil
}

// CallIntInt calls `name` function.
//  argument type: int
//  return type:   int
func (m *ObjectModule) CallIntInt(name string, i int) (int, error) {
	pyFunc, err := m.getPyFunc(name)
	if err != nil {
		return 0, fmt.Errorf("%v at '%v'", err.Error(), name)
	}
	defer pyFunc.DecRef()

	pyArg := C.PyTuple_New(1)
	defer C.Py_DecRef(pyArg)
	pyValue := C.PyInt_FromLong(C.long(i))
	defer C.Py_DecRef(pyValue)
	C.PyTuple_SetItem(pyArg, 0, pyValue)

	ret, err := pyFunc.CallObject(Object{p: pyArg})
	if ret.p == nil && err != nil {
		return 0, fmt.Errorf("%v in '%v'", err.Error(), name)
	}

	return int(C.PyInt_AsLong(ret.p)), nil
}

// CallNoneString calls `name` function.
//  argument:    nothing
//  return type: string
func (m *ObjectModule) CallNoneString(name string) (string, error) {
	pyFunc, err := m.getPyFunc(name)
	if err != nil {
		return "", fmt.Errorf("%v at '%v'", err.Error(), name)
	}
	defer pyFunc.DecRef()

	ret, err := pyFunc.CallObject(Object{})
	if ret.p == nil && err != nil {
		return "", fmt.Errorf("%v in '%v'", err.Error(), name)
	}

	charPtr := C.PyString_AsString(ret.p)
	return C.GoString(charPtr), nil
}

// CallNone2String calls `name` function.
//  argument:    nothing
//  return type: two strings
func (m *ObjectModule) CallNone2String(name string) (string, string, error) {
	pyFunc, err := m.getPyFunc(name)
	if err != nil {
		return "", "", fmt.Errorf("%v at '%v'", err.Error(), name)
	}
	defer pyFunc.DecRef()

	ret, err := pyFunc.CallObject(Object{})
	if ret.p == nil && err != nil {
		return "", "", fmt.Errorf("%v in '%v'", err.Error(), name)
	}

	r1 := C.PyTuple_GetItem(ret.p, 0)
	if r1 == nil {
		return "", "", fmt.Errorf("cannot get 1st. result")
	}
	defer C.Py_DecRef(r1)

	r2 := C.PyTuple_GetItem(ret.p, 1)
	if r2 == nil {
		return "", "", fmt.Errorf("cannot get 2nd. result")
	}
	defer C.Py_DecRef(r2)

	charPtr1 := C.PyString_AsString(r1)
	charPtr2 := C.PyString_AsString(r2)
	return C.GoString(charPtr1), C.GoString(charPtr2), nil
}

// CallStringString calls `name` function.
//  argument type: string
//  return type:   string
func (m *ObjectModule) CallStringString(name string, s string) (string, error) {
	pyFunc, err := m.getPyFunc(name)
	if err != nil {
		return "", fmt.Errorf("%v at '%v'", err.Error(), name)
	}
	defer pyFunc.DecRef()

	pyArg := C.PyTuple_New(1)
	defer C.Py_DecRef(pyArg)
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	pyValue := C.PyString_FromString(cs) // this pyValue is not need to DECREF
	C.PyTuple_SetItem(pyArg, 0, pyValue)

	ret, err := pyFunc.CallObject(Object{p: pyArg})
	if ret.p == nil && err != nil {
		return "", fmt.Errorf("%v in '%v'", err.Error(), name)
	}

	charPtr := C.PyString_AsString(ret.p)
	return C.GoString(charPtr), nil
}

// CallByteByte calls `name` function.
//  argument type: []byte (byteArray in python)
//  return type:   []byte
func (m *ObjectModule) CallByteByte(name string, b []byte) ([]byte, error) {
	pyFunc, err := m.getPyFunc(name)
	if err != nil {
		return []byte{}, fmt.Errorf("%v at '%v'", err.Error(), name)
	}
	defer pyFunc.DecRef()

	pyArg := C.PyTuple_New(1)
	defer C.Py_DecRef(pyArg)
	cb := (*C.char)(unsafe.Pointer(&b[0]))
	pyByte := C.PyByteArray_FromStringAndSize(cb, C.Py_ssize_t(len(b)))
	C.PyTuple_SetItem(pyArg, 0, pyByte)

	ret, err := pyFunc.CallObject(Object{p: pyArg})
	if ret.p == nil && err != nil {
		return []byte{}, fmt.Errorf("%v in '%v'", err.Error(), name)
	}

	bytePtr := C.PyByteArray_FromObject(ret.p)
	charPtr := C.PyByteArray_AsString(bytePtr)
	l := C.PyByteArray_Size(ret.p)
	return C.GoBytes(unsafe.Pointer(charPtr), C.int(l)), nil
}

// Call calls `name` function.
//  argument type: ...data.Value
//  return type:   data.Value
func (m *ObjectModule) Call(name string, args ...data.Value) (data.Value, error) {
	var res data.Value
	pyFunc, err := m.getPyFunc(name)
	if err != nil {
		return res, fmt.Errorf("%v at '%v'", err.Error(), name)
	}
	defer pyFunc.DecRef()

	pyArg := C.PyTuple_New(C.Py_ssize_t(len(args)))
	defer C.Py_DecRef(pyArg)

	for i, v := range args {
		o := newPyObj(v)
		C.PyTuple_SetItem(pyArg, C.Py_ssize_t(i), o.p)
	}

	ret, err := pyFunc.CallObject(Object{p: pyArg})
	if ret.p == nil && err != nil {
		return res, fmt.Errorf("%v in '%v'", err.Error(), name)
	}

	return fromPyTypeObject(ret.p), nil
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

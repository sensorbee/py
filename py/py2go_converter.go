package py

/*
#cgo darwin pkg-config: python-2.7
#include "Python.h"
*/
import "C"
import (
	"pfi/sensorbee/sensorbee/data"
	"unsafe"
)

func fromPyTypeObject(o *C.PyObject) data.Value {
	switch {
	// FIXME: this internal code should not use
	case o == (*C.PyObject)(unsafe.Pointer(&C._Py_TrueStruct)):
		return data.Bool(true)

	// FIXME: this internal code should not use
	case o == (*C.PyObject)(unsafe.Pointer(&C._Py_ZeroStruct)):
		return data.Bool(false)

	case o.ob_type == &C.PyInt_Type:
		return data.Int(C.PyInt_AsLong(o))

	case o.ob_type == &C.PyFloat_Type:
		return data.Float(C.PyFloat_AsDouble(o))

	case o.ob_type == &C.PyByteArray_Type:
		bytePtr := C.PyByteArray_FromObject(o)
		charPtr := C.PyByteArray_AsString(bytePtr)
		l := C.PyByteArray_Size(o)
		return data.Blob(C.GoBytes(unsafe.Pointer(charPtr), C.int(l)))

	case o.ob_type == &C.PyString_Type:
		size := C.int(C.PyString_Size(o))
		charPtr := C.PyString_AsString(o)
		return data.String(string(C.GoBytes(unsafe.Pointer(charPtr), size)))

	case o.ob_type == &C.PyList_Type:
		return fromPyArray(o)

	case o.ob_type == &C.PyDict_Type:
		return fromPyMap(o)

	// FIXME: this internal code should not use
	case o == &C._Py_NoneStruct:
		return data.Null{}

	}
	// TODO: implement timestamp

	return data.Null{}
}

func fromPyArray(ls *C.PyObject) data.Array {
	size := int(C.PyList_Size(ls))
	array := make(data.Array, size)
	for i := 0; i < size; i++ {
		o := C.PyList_GetItem(ls, C.Py_ssize_t(i))
		array[i] = fromPyTypeObject(o)
	}
	return array
}

func fromPyMap(o *C.PyObject) data.Map {
	m := data.Map{}

	var key, value *C.PyObject
	pos := C.Py_ssize_t(C.int(0))

	for C.int(C.PyDict_Next(o, &pos, &key, &value)) > 0 {
		// data.Map's key is only allowed string
		if key.ob_type != &C.PyString_Type {
			continue
		}
		key, _ := data.ToString(fromPyTypeObject(key))
		m[key] = fromPyTypeObject(value)
	}

	return m
}

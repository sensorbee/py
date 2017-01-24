// +build py3.4 py3.5 py3.6

package py

/*
#include "Python.h"

int IsPyTypeTrue(PyObject *o) {
  return o == Py_True;
}

int IsPyTypeFalse(PyObject *o) {
  return o == Py_False;
}

int IsPyTypeLong(PyObject *o) {
  return PyLong_CheckExact(o);
}

int IsPyTypeFloat(PyObject *o) {
  return PyFloat_CheckExact(o);
}

int IsPyTypeByteArray(PyObject *o) {
  return PyByteArray_CheckExact(o);
}

int IsPyTypeBytes(PyObject *o) {
  return PyBytes_CheckExact(o);
}

int IsPyTypeList(PyObject *o) {
  return PyList_CheckExact(o);
}

int IsPyTypeDict(PyObject *o) {
  return PyDict_CheckExact(o);
}

int IsPyTypeTuple(PyObject *o) {
  return PyTuple_CheckExact(o);
}

int IsPyTypeNone(PyObject *o) {
  return o == Py_None;
}

PyTypeObject* GetTypeObject(PyObject *o) {
  return (PyTypeObject*)PyObject_Type(o);
}

const char* GetTypeName(PyTypeObject *t) {
  return t->tp_name;
}

void DecRefTypeObject(PyTypeObject *t) {
  Py_DECREF(t);
}
*/
import "C"
import (
	"fmt"
	"unsafe"

	"gopkg.in/sensorbee/sensorbee.v0/data"
)

func fromPyTypeObject(o *C.PyObject) (data.Value, error) {
	switch {
	case C.IsPyTypeTrue(o) > 0:
		return data.Bool(true), nil

	case C.IsPyTypeFalse(o) > 0:
		return data.Bool(false), nil

	case C.IsPyTypeLong(o) > 0:
		return data.Int(C.PyLong_AsLong(o)), nil

	case C.IsPyTypeFloat(o) > 0:
		return data.Float(C.PyFloat_AsDouble(o)), nil

	case C.IsPyTypeByteArray(o) > 0:
		bytePtr := C.PyByteArray_FromObject(o)
		charPtr := C.PyByteArray_AsString(bytePtr)
		size := C.int(C.PyByteArray_Size(o))
		return data.Blob(C.GoBytes(unsafe.Pointer(charPtr), size)), nil

	case C.IsPyTypeBytes(o) > 0:
		charPtr := C.PyBytes_AsString(o)
		size := C.int(C.PyBytes_Size(o))
		return data.Blob(C.GoBytes(unsafe.Pointer(charPtr), size)), nil

	case isPyTypeUnicode(o) > 0:
		// Use unicode string as UTF-8 in py because
		// Go's source code is defined to be UTF-8 text and string literal is too.
		var size C.Py_ssize_t
		charPtr := C.PyUnicode_AsUTF8AndSize(o, &size)
		if charPtr == nil {
			return data.Null{}, getPyErr()
		}
		return data.String(string(C.GoBytes(unsafe.Pointer(charPtr), C.int(size)))), nil

	case isPyTypeDateTime(o):
		return fromTimestamp(o), nil

	case C.IsPyTypeList(o) > 0:
		return fromPyArray(o)

	case C.IsPyTypeDict(o) > 0:
		return fromPyMap(o)

	case C.IsPyTypeTuple(o) > 0:
		return fromPyTuple(o)

	case C.IsPyTypeNone(o) > 0:
		return data.Null{}, nil

	}

	t := C.GetTypeObject(o)
	if t == nil {
		return data.Null{}, fmt.Errorf("unsupported type in sensorbee/py (cannot detect python object type)")
	}
	tn := C.GoString(C.GetTypeName(t))
	defer C.DecRefTypeObject(t)
	return data.Null{}, fmt.Errorf("unsupported type in sensorbee/py: %v", tn)
}

func isPyTypeString(o *C.PyObject) int {
	return int(C.IsPyTypeBytes(o))
}

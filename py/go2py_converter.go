package py

/*
#cgo pkg-config: python-2.7
#include "Python.h"
*/
import "C"
import (
	"pfi/sensorbee/sensorbee/data"
	"unsafe"
)

func getNewPyDic(m map[string]interface{}) Object {
	return Object{}
}

func newPyObj(v data.Value) Object {
	var pyobj *C.PyObject
	switch v.Type() {
	case data.TypeBool:
		b, _ := data.ToInt(v)
		pyobj = C.PyBool_FromLong(C.long(b))
	case data.TypeInt:
		i, _ := data.AsInt(v)
		pyobj = C.PyInt_FromLong(C.long(i))
	case data.TypeFloat:
		f, _ := data.AsFloat(v)
		pyobj = C.PyFloat_FromDouble(C.double(f))
	case data.TypeString:
		s, _ := data.AsString(v)
		pyobj = newPyString(s)
	case data.TypeBlob:
		b, _ := data.AsBlob(v)
		cb := (*C.char)(unsafe.Pointer(&b[0]))
		pyobj = C.PyByteArray_FromStringAndSize(cb, C.Py_ssize_t(len(b)))
	case data.TypeTimestamp:
		t, _ := data.AsTimestamp(v)
		pyobj = getPyDateTime(t)
	case data.TypeArray:
		innerArray, _ := data.AsArray(v)
		pyobj = newPyArray(innerArray)
	case data.TypeMap:
		innerMap, _ := data.AsMap(v)
		pyobj = newPyMap(innerMap)
	case data.TypeNull:
		// FIXME: this internal code should not use
		pyobj = &C._Py_NoneStruct
		pyobj.ob_refcnt++
	default:
		// TODO: change error
		panic("not implemented!")
	}
	return Object{p: pyobj}
}

func newPyString(s string) *C.PyObject {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return C.PyString_FromString(cs)
}

func newPyArray(a data.Array) *C.PyObject {
	pylist := C.PyList_New(C.Py_ssize_t(len(a)))
	for i, v := range a {
		value := newPyObj(v)
		// PyList object takes over the value's reference, and not need to
		// decrease reference counter.
		C.PyList_SetItem(pylist, C.Py_ssize_t(i), value.p)
	}
	return pylist
}

func newPyMap(m data.Map) *C.PyObject {
	pydict := C.PyDict_New()
	for k, v := range m {
		func() {
			key := newPyString(k)
			value := newPyObj(v)
			defer value.decRef()
			C.PyDict_SetItem(pydict, key, value.p)
		}()
	}
	return pydict
}

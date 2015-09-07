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

func getNewPyDic(m map[string]interface{}) Object {
	return Object{}
}

func newPyObj(v data.Value) Object {
	var pyobj *C.PyObject
	defer C.Py_DecRef(pyobj)
	switch v.Type() {
	case data.TypeBool:
		panic("not implemented!")
	case data.TypeInt:
		i, _ := data.AsInt(v)
		pyobj = C.PyInt_FromLong(C.long(i))
	case data.TypeFloat:
		panic("not implemented!")
	case data.TypeString:
		s, _ := data.AsString(v)
		pyobj = newPyString(s)
	case data.TypeBlob:
		b, _ := data.AsBlob(v)
		cb := (*C.char)(unsafe.Pointer(&b[0]))
		pyobj = C.PyByteArray_FromStringAndSize(cb, C.Py_ssize_t(len(b)))
	case data.TypeTimestamp:
		panic("not implemented!")
	case data.TypeArray:
		panic("not implemented!")
	case data.TypeMap:
		innerMap, _ := data.AsMap(v)
		pyobj = newPyMap(innerMap)
	case data.TypeNull:
		panic("not implemented!")
	default:
		panic("not implemented!")
	}
	return Object{p: pyobj}
}

func newPyString(s string) *C.PyObject {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return C.PyString_FromString(cs)
}

func newPyArray(a data.Array) []interface{} {
	result := make([]interface{}, len(a))
	for i, v := range a {
		value := newPyObj(v)
		result[i] = value
	}
	return result
}

func newPyMap(m data.Map) *C.PyObject {
	pydict := C.PyDict_New()
	for k, v := range m {
		key := newPyString(k)
		value := newPyObj(v)
		C.PyDict_SetItem(pydict, key, value.p)
	}
	return pydict
}

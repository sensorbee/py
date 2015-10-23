package py

/*
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"pfi/sensorbee/sensorbee/data"
	"unsafe"
)

func getNewPyDic(m map[string]interface{}) Object {
	return Object{}
}

func newPyObj(v data.Value) (Object, error) {
	var pyobj *C.PyObject
	var err error
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
		pyobj, err = newPyArray(innerArray)
	case data.TypeMap:
		innerMap, _ := data.AsMap(v)
		pyobj, err = newPyMap(innerMap)
	case data.TypeNull:
		// FIXME: this internal code should not use
		pyobj = &C._Py_NoneStruct
		pyobj.ob_refcnt++
	default:
		err = fmt.Errorf("not supported type in pystate: %s", v.Type())
	}
	return Object{p: pyobj}, err
}

func newPyString(s string) *C.PyObject {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return C.PyString_FromString(cs)
}

func newPyArray(a data.Array) (*C.PyObject, error) {
	pylist := C.PyList_New(C.Py_ssize_t(len(a)))
	for i, v := range a {
		value, err := newPyObj(v)
		if err != nil {
			return nil, err
		}
		// PyList object takes over the value's reference, and not need to
		// decrease reference counter.
		C.PyList_SetItem(pylist, C.Py_ssize_t(i), value.p)
	}
	return pylist, nil
}

func newPyMap(m data.Map) (*C.PyObject, error) {
	pydict := C.PyDict_New()
	for k, v := range m {
		err := func() error {
			key := newPyString(k)
			value, err := newPyObj(v)
			if err != nil {
				return err
			}
			defer value.decRef()
			C.PyDict_SetItem(pydict, key, value.p)
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return pydict, nil
}

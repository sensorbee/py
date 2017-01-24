package py

/*
#include "Python.h"

PyObject* getPyNone() {
  return Py_BuildValue("");
}
*/
import "C"
import (
	"unsafe"

	"gopkg.in/sensorbee/sensorbee.v0/data"
)

func getPyNone() *C.PyObject {
	return C.getPyNone()
}

func getNewPyDic(m map[string]interface{}) Object {
	return Object{}
}

func newPyString(s string) *C.PyObject {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return C.PyUnicode_FromString(cs)
}

func newPyArray(a data.Array) (*C.PyObject, error) {
	pylist := C.PyList_New(C.Py_ssize_t(len(a)))
	if pylist == nil {
		return nil, getPyErr()
	}
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
	if pydict == nil {
		return nil, getPyErr()
	}
	for k, v := range m {
		err := func() error {
			ck := C.CString(k)
			defer C.free(unsafe.Pointer(ck))
			value, err := newPyObj(v)
			if err != nil {
				return err
			}
			defer value.decRef()
			C.PyDict_SetItemString(pydict, ck, value.p)
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return pydict, nil
}

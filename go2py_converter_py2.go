// +build !py3.4
// +build !py3.5
// +build !py3.6

package py

/*
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"unsafe"

	"gopkg.in/sensorbee/sensorbee.v0/data"
)

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
		if len(b) == 0 {
			pyobj = C.PyByteArray_FromStringAndSize(nil, 0)
		} else {
			cb := (*C.char)(unsafe.Pointer(&b[0]))
			pyobj = C.PyByteArray_FromStringAndSize(cb, C.Py_ssize_t(len(b)))
		}
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
		pyobj = getPyNone()
	default:
		err = fmt.Errorf("unsupported type in sensorbee/py: %s", v.Type())
	}

	if pyobj == nil && err == nil {
		return Object{}, getPyErr()
	}
	return Object{p: pyobj}, err
}

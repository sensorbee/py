package py

/*
#cgo darwin pkg-config: python-2.7
#include "Python.h"
*/
import "C"

// Object is a bind of `*C.PyObject`
type Object struct {
	p *C.PyObject
}

// DecRef declare reference counter of `C.PyObject`
func (o *Object) DecRef() {
	C.Py_DecRef(o.p)
}

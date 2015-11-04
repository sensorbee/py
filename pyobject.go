package py

/*
#include "Python.h"
void XDecRef(PyObject *o) {
  Py_XDECREF(o);
}
*/
import "C"
import (
	"pfi/sensorbee/py/mainthread"
)

// Object is a bind of `*C.PyObject`
type Object struct {
	p *C.PyObject
}

// DecRef decrease reference counter of `C.PyObject`
// This function is public for API users and
// it acquires GIL of Python interpreter.
// A user can safely call this method even when its target object is null.
func (o *Object) DecRef() {
	mainthread.ExecSync(func() { // TODO: This Exec should probably be removed.
		C.XDecRef(o.p)
	})
}

// decRef decrease reference counter of `C.PyObject`
// This function doesn't acquire GIL.
func (o *Object) decRef() {
	C.Py_DecRef(o.p)
}

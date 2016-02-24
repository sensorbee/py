package py

/*
#include "Python.h"
*/
import "C"
import (
	"gopkg.in/sensorbee/py.v0/mainthread"
)

// Object is a bind of `*C.PyObject`
type Object struct {
	p *C.PyObject
}

// Release decreases reference counter of `C.PyObject` and released the object.
// This function is public for API users and it acquires GIL of Python
// interpreter. A user can safely call this method even when its target object
// is null.
func (o *Object) Release() {
	mainthread.ExecSync(func() { // TODO: This Exec should probably be removed.
		// Py_XDECREF is not used here because it causes SEGV on Windows.
		if o.p == nil {
			return
		}
		C.Py_DecRef(o.p)
		o.p = nil
	})
}

// decRef decrease reference counter of `C.PyObject`
// This function doesn't acquire GIL.
func (o *Object) decRef() {
	C.Py_DecRef(o.p)
}

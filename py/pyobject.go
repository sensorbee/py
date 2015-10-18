package py

/*
#include "Python.h"
void XDecRef(PyObject *o) {
  Py_XDECREF(o);
}
*/
import "C"
import (
	"runtime"
)

// Object is a bind of `*C.PyObject`
type Object struct {
	p *C.PyObject
}

// DecRef decrease reference counter of `C.PyObject`
// This function is public for API users and
// it acquires GIL of Python interpreter.
// User can call whenever target object is null.
func (o *Object) DecRef() {
	ch := make(chan bool, 1)
	go func() {
		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)

		C.XDecRef(o.p)
		ch <- true
	}()
	<-ch
}

// decRef decrease reference counter of `C.PyObject`
// This function doesn't acquire GIL.
func (o *Object) decRef() {
	C.Py_DecRef(o.p)
}

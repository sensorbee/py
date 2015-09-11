package py

/*
#cgo darwin pkg-config: python-2.7
#include "Python.h"
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
func (o *Object) DecRef() {
	ch := make(chan bool, 1)
	go func() {
		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)

		C.Py_DecRef(o.p)
		ch <- true
	}()
	<-ch
}

// decRef decrease reference counter of `C.PyObject`
// This function doesn't acquire GIL.
func (o *Object) decRef() {
	C.Py_DecRef(o.p)
}

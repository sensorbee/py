package py

/*
#cgo darwin pkg-config: python-2.7
#include "Python.h"
*/
import "C"
import (
	"fmt"
)

// ObjectFunc is a bind of `PyObject` used as `PyFunc`
type ObjectFunc struct {
	Object
}

// CallObject executes python function, using `PyObject_CallObject`. Returns a
// `PyObject` even if result values are more thane one. When a value will be set
// directory, and values will be set as a `PyTuple` object.
func (f *ObjectFunc) CallObject(arg Object) (po Object, err error) {
	po = Object{}
	pyValue, err := C.PyObject_CallObject(f.p, arg.p)
	if pyValue == nil && err != nil {
		err = fmt.Errorf("call function error: %v", err)
	} else {
		po.p = pyValue
	}
	return po, err
}

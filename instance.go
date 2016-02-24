package py

/*
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"gopkg.in/sensorbee/py.v0/mainthread"
	"gopkg.in/sensorbee/sensorbee.v0/data"
	"unsafe"
)

// ObjectInstance is a bind of Python instance, used as `PyInstance`.
type ObjectInstance struct {
	Object
}

// Call calls `name` function.
// [TODO] this function is not supported named arguments
func (ins *ObjectInstance) Call(name string, args ...data.Value) (data.Value,
	error) {
	type Result struct {
		val data.Value
		err error
	}
	ch := make(chan *Result)
	mainthread.Exec(func() {
		if ins.p == nil {
			ch <- &Result{nil, fmt.Errorf("ins.p of %p is nil while calling %s", ins, name)}
			return
		}
		v, err := invoke(ins.p, name, args, nil)
		ch <- &Result{v, err}
	})
	res := <-ch
	return res.val, res.err
}

// CallDirect calls `name` function and return `PyObject` directly.
// This method is suitable for getting the instance object that called method
// returned.
func (ins *ObjectInstance) CallDirect(name string, args []data.Value,
	kwdArg data.Map) (Object, error) {
	type Result struct {
		val Object
		err error
	}
	ch := make(chan *Result)
	mainthread.Exec(func() {
		v, err := invokeDirect(ins.p, name, args, kwdArg)
		ch <- &Result{v, err}
	})
	res := <-ch
	return res.val, res.err
}

func newInstance(m *ObjectModule, name string, args []data.Value, kwdArgs data.Map) (
	result ObjectInstance, resErr error) {

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	defer func() {
		if r := recover(); r != nil {
			resErr = fmt.Errorf("cannot call '%v' due to panic: %v", name, r)
		}
	}()

	pyInstance := C.PyObject_GetAttrString(m.p, cName)
	if pyInstance == nil {
		return ObjectInstance{}, fmt.Errorf("fail to get '%v' class: %v", name, getPyErr())
	}
	defer C.Py_DecRef(pyInstance)

	// no named arguments
	pyArg, err := convertArgsGo2Py(args)
	if err != nil {
		return ObjectInstance{}, fmt.Errorf("fail to convert non named arguments in creating '%v' instance: %v",
			name, err.Error())
	}
	defer pyArg.decRef()

	// named arguments
	var pyKwdArg *C.PyObject
	if len(kwdArgs) == 0 {
		pyKwdArg = nil
	} else {
		o, err := newPyObj(kwdArgs)
		if err != nil {
			return ObjectInstance{}, fmt.Errorf("fail to convert named arguments in creating '%v' instance: %v",
				name, err.Error())
		}
		defer o.decRef()
		pyKwdArg = o.p
	}

	ret := C.PyObject_Call(pyInstance, pyArg.p, pyKwdArg)
	if ret == nil {
		return ObjectInstance{}, fmt.Errorf("fail to create '%v' instance: %v", name, getPyErr())
	}

	return ObjectInstance{Object{p: ret}}, nil
}

// +build !py3.4
// +build !py3.5
// +build !py3.6

package py

/*
#include "Python.h"

PyObject* idOrNone(PyObject* o)
{
  return o ? o : Py_BuildValue("");
}

void fetchPythonError(PyObject* excInfo)
{
  PyObject *type, *value, *traceback;
  PyErr_Fetch(&type, &value, &traceback);
  PyTuple_SetItem(excInfo, 0, idOrNone(type));
  PyTuple_SetItem(excInfo, 1, idOrNone(value));
  PyTuple_SetItem(excInfo, 2, idOrNone(traceback));
}
*/
import "C"
import (
	"errors"
	"unsafe"

	"gopkg.in/sensorbee/py.v0/mainthread"
)

func init() {
	ch := make(chan error)
	mainthread.Exec(func() {
		traceback, err := loadModule("traceback")
		if err != nil {
			ch <- err
			return
		}
		defer traceback.decRef()
		formatException, err := getPyFunc(traceback.p, "format_exception")
		if err != nil {
			ch <- err
			return
		}
		tracebackFormatExceptionFunc = formatException

		exceptions, err := loadModule("exceptions")
		if err != nil {
			ch <- err
			return
		}
		defer exceptions.decRef()
		syntaxErrorCString := C.CString("SyntaxError")
		defer C.free(unsafe.Pointer(syntaxErrorCString))
		syntaxError := C.PyObject_GetAttrString(exceptions.p, syntaxErrorCString)
		if syntaxError == nil {
			ch <- errors.New("cannot load exceptions.SyntaxError")
			return
		}
		syntaxErrorType.p = syntaxError

		ch <- nil
	})

	if err := <-ch; err != nil {
		panic(err)
	}
}

func fetchPythonError(o Object) {
	C.fetchPythonError(o.p)
}

func extractLineFromFormattedErrorMessage(formatted Object, n C.Py_ssize_t) string {
	line := C.PyList_GetItem(formatted.p, n)
	return C.GoString(C.PyString_AsString(line))
}

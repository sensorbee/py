package p

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
	"fmt"
	"runtime"
	"strings"
	"unicode"
	"unsafe"
)

func init() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	releaseGIL, err := initAndLockPython()
	if err != nil {
		panic(err)
	}
	defer releaseGIL()

	traceback, err := loadModule("traceback")
	if err != nil {
		panic(err)
	}
	defer traceback.decRef()
	if formatException, err := getPyFunc(traceback.p, "format_exception"); err != nil {
		panic(err)
	} else {
		tracebackFormatExceptionFunc = formatException
	}

	exceptions, err := loadModule("exceptions")
	if err != nil {
		panic(err)
	}
	defer exceptions.decRef()
	syntaxErrorCString := C.CString("SyntaxError")
	defer C.free(unsafe.Pointer(syntaxErrorCString))
	if syntaxError := C.PyObject_GetAttrString(exceptions.p, syntaxErrorCString); syntaxError == nil {
		panic("cannot load exceptions.SyntaxError")
	} else {
		syntaxErrorType.p = syntaxError
	}
}

func loadModule(name string) (mod ObjectModule, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if m := C.PyImport_ImportModule(cName); m == nil {
		err = fmt.Errorf("failed to import '%v'", name)
	} else {
		mod.p = m
	}
	return
}

var tracebackFormatExceptionFunc ObjectFunc
var syntaxErrorType Object

// tracebackFormatException calls traceback.format_exception().
func tracebackFormatException(excInfo Object) (Object, error) {
	formatted := C.PyObject_CallObject(tracebackFormatExceptionFunc.p, excInfo.p)
	if formatted == nil {
		return Object{p: nil}, errors.New("failed to extract error info from Python")
	}
	return Object{p: formatted}, nil
}

func isSyntaxError(p *C.PyObject) bool {
	return C.PyType_IsSubtype(pyObjectToPyTypeObject(p), pyObjectToPyTypeObject(syntaxErrorType.p)) != 0
}

func pyObjectToPyTypeObject(p *C.PyObject) *C.PyTypeObject {
	return (*C.PyTypeObject)(unsafe.Pointer(p))
}

func getPyErr() error {
	if isPyNoMemoryError() {
		// Fetching stacktrace requires some memory,
		// so just return an error without stacktrace.
		return pyNoMemoryError
	}

	// TODO: consider to reserve excInfo
	excInfo := Object{p: C.PyTuple_New(3)}
	if excInfo.p == nil {
		if isPyNoMemoryError() {
			return pyNoMemoryError
		}
		return getPyErr()
	}
	defer excInfo.decRef()
	C.fetchPythonError(excInfo.p)

	formatted, err := tracebackFormatException(excInfo)
	if err != nil {
		return err
	}
	defer formatted.decRef()

	ln := C.PyList_Size(formatted.p)
	mainMsg := strings.TrimSpace(extractLineFromFormattedErrorMessage(formatted, ln-1))
	syntaxErr := ""
	nTracebackLines := ln - 1
	if isSyntaxError(C.PyTuple_GetItem(excInfo.p, 0)) {
		firstLine := extractLineFromFormattedErrorMessage(formatted, ln-3)
		secondLine := extractLineFromFormattedErrorMessage(formatted, ln-2)
		syntaxErr = firstLine + secondLine
		nTracebackLines = ln - 3
	}
	stackTrace := ""
	for i := C.Py_ssize_t(0); i < nTracebackLines; i++ {
		stackTrace += extractLineFromFormattedErrorMessage(formatted, i)
	}
	stackTrace = strings.TrimRightFunc(stackTrace, unicode.IsSpace)
	return &pyErr{
		mainMsg:      mainMsg,
		syntaxErrMsg: syntaxErr,
		stackTrace:   stackTrace,
	}
}

func extractLineFromFormattedErrorMessage(formatted Object, n C.Py_ssize_t) string {
	line := C.PyList_GetItem(formatted.p, n)
	return C.GoString(C.PyString_AsString(line))
}

type pyErr struct {
	mainMsg      string
	syntaxErrMsg string
	stackTrace   string
}

// Error returns an error message string for pyErr.
// This string contains multiple lines for stacktrace.
func (e *pyErr) Error() string {
	return e.mainMsg + "\n" + e.syntaxErrMsg + e.stackTrace
}

func isPyNoMemoryError() bool {
	return C.PyErr_ExceptionMatches(C.PyExc_MemoryError) != 0
}

// pyNoMemoryError is an error value representing an allocation error on Python.
// This is not a pyErr because it is difficult to extract stacktrace from Python when the heap is exhaused.
var pyNoMemoryError = errors.New("python interpreter failed to allocate memory")

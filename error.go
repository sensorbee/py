package py

/*
#include "Python.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unsafe"
)

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

// getPyErr returns a Python's exception as an error.
// getPyErr normally returns a pyErr (see its godoc for details) and clears
// exception state of the python interpreter. If the exception is a MemoryError,
// getPyErr returns pyNoMemoryError (without stacktrace) and does not clears
// the exception state. The easiest way to extract stacktrace for MemoryError
// is calling PyErr_Print(). Note that PyErr_Print() prints to stderr.
func getPyErr() error {
	if isPyNoMemoryError() {
		// Fetching stacktrace requires some memory,
		// so just return an error without stacktrace.
		return errPyNoMemory
	}

	// TODO: consider to reserve excInfo
	excInfo := Object{p: C.PyTuple_New(3)}
	if excInfo.p == nil {
		if isPyNoMemoryError() {
			return errPyNoMemory
		}
		return getPyErr()
	}
	defer excInfo.decRef()
	fetchPythonError(excInfo)

	formatted, err := tracebackFormatException(excInfo)
	if err != nil {
		return err
	}
	defer formatted.decRef()

	ln := C.PyList_Size(formatted.p)
	mainMsg := strings.TrimSpace(extractLineFromFormattedErrorMessage(formatted, ln-1))
	syntaxErr := ""
	nTracebackLines := ln - 1
	if isSyntaxError(C.PyTuple_GetItem(excInfo.p, 0)) && ln > 2 {
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

// pyErr represents an exception of python.
type pyErr struct {
	mainMsg      string // "main error message" (one line)
	syntaxErrMsg string // syntax error description for SyntaxError (zero or two lines)
	stackTrace   string // stacktrace (zero or multiple lines)
}

// Error returns an error message string for pyErr.
// This string contains multiple lines for stacktrace.
func (e *pyErr) Error() string {
	return e.mainMsg + "\n" + e.syntaxErrMsg + e.stackTrace
}

func isPyNoMemoryError() bool {
	return C.PyErr_ExceptionMatches(C.PyExc_MemoryError) != 0
}

// errPyNoMemory is an error value representing an allocation error on Python.
// This is not a pyErr because it is difficult to extract stacktrace from Python when the heap is exhaused.
var errPyNoMemory = errors.New("python interpreter failed to allocate memory")

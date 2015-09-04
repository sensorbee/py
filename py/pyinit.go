package py

/*
#cgo darwin pkg-config: python-2.7
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Initialize python interpreter and GIL.
func Initialize() error {
	if C.Py_IsInitialized() == 0 {
		C.Py_Initialize()
	}

	if C.Py_IsInitialized() == 0 {
		return fmt.Errorf("cannot initialize python command")
	}

	if C.PyEval_ThreadsInitialized() == 0 {
		C.PyEval_InitThreads()
	}

	if C.PyEval_ThreadsInitialized() == 0 {
		return fmt.Errorf("cannot initialize GIL")
	}

	return nil
}

// ImportSysAndAppendPath sets `sys.path` to load modules.
func ImportSysAndAppendPath(paths ...string) error {
	importSys := C.CString("import sys")
	defer C.free(unsafe.Pointer(importSys))
	C.PyRun_SimpleStringFlags(importSys, nil)

	for _, path := range paths {
		sysPath := fmt.Sprintf("sys.path.append('%v')", path)
		cSysPath := C.CString(sysPath)
		defer C.free(unsafe.Pointer(cSysPath))
		C.PyRun_SimpleStringFlags(cSysPath, nil)
	}

	return nil
}

// Finalize python interpreter. Attention that the process does not collect all
// object. User need to implement that declare reference count manually when the
// object is not managed by the interpreter.
func Finalize() error {
	// when not initialized, should not finalize but no problem
	C.Py_Finalize()
	return nil
}

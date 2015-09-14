package py

/*
#cgo pkg-config: python-2.7
#include "Python.h"

int GILStateEnsure() {
	return PyGILState_Ensure();
}

void GILStateRelease(int gstate) {
	PyGILState_Release(gstate);
}
*/
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

// initialize python interpreter and GIL.
func initialize() error {
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

func GILState_Ensure() int {
	return int(C.GILStateEnsure())
}

func GILState_Release(gstate int) {
	C.GILStateRelease(C.int(gstate))
}

func LockGILAndExecute(f func()) {
	ch := make(chan bool, 1)
	go func() {
		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)
		f()
		ch <- true
	}()
	<-ch
}

// ImportSysAndAppendPath sets `sys.path` to load modules.
func ImportSysAndAppendPath(paths ...string) error {
	LockGILAndExecute(func() {
		importSys := C.CString("import sys")
		defer C.free(unsafe.Pointer(importSys))
		C.PyRun_SimpleStringFlags(importSys, nil)

		for _, path := range paths {
			sysPath := fmt.Sprintf("sys.path.append('%v')", path)
			cSysPath := C.CString(sysPath)
			defer C.free(unsafe.Pointer(cSysPath))
			C.PyRun_SimpleStringFlags(cSysPath, nil)
		}
	})

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

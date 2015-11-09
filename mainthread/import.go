package mainthread

/*
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func importSys() error {
	importSys := C.CString("import sys")
	defer C.free(unsafe.Pointer(importSys))
	if result := C.PyRun_SimpleStringFlags(importSys, nil); result != 0 {
		return fmt.Errorf(`fail to import "sys" package`)
	}
	return nil
}

// AppendSysPath sets `sys.path` to load modules.
func AppendSysPath(path string) error {
	ch := make(chan error)
	Exec(func() { // TODO: This Exec should probably be provided out side this function.
		sysPath := fmt.Sprintf("sys.path.append('%v')", path)
		cSysPath := C.CString(sysPath)
		defer C.free(unsafe.Pointer(cSysPath))
		if result := C.PyRun_SimpleStringFlags(cSysPath, nil); result != 0 {
			ch <- fmt.Errorf("fail to append '%v' path", path)
			return
		}
		ch <- nil
	})
	err := <-ch
	return err
}

// AppendSysPathNoGIL sets `sys.path` to load modules.
func AppendSysPathNoGIL(path string) error {
	sysPath := fmt.Sprintf("sys.path.append('%v')", path)
	cSysPath := C.CString(sysPath)
	defer C.free(unsafe.Pointer(cSysPath))
	if result := C.PyRun_SimpleStringFlags(cSysPath, nil); result != 0 {
		return fmt.Errorf("fail to append '%v' path", path)
	}

	return nil
}

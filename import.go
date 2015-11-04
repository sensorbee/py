package py

/*
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"pfi/sensorbee/py/mainthread"
	"unsafe"
)

// ImportSysAndAppendPath sets `sys.path` to load modules.
func ImportSysAndAppendPath(paths ...string) error {
	mainthread.Exec(func() { // TODO: This Exec should probably be provided out side this function.
		// TODO: All import errors must be detected
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

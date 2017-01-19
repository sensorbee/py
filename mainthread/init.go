/*
Package mainthread initializes Python interpreter and lock the goroutine to a
specific OS thread so that the interpreter can keep using the same OS thread
as a main thread. It also provides Exec function to run any function on the
main thread.

This is provided as a separate package to guarantee that the init function in
this package is executed before all other init functions in py package.
*/
package mainthread

/*
#include "Python.h"
*/
import "C"
import (
	"errors"
	"runtime"
)

func init() {
	ch := make(chan error)
	go func() {
		// Python interpreter needs to run on the same OS thread.
		runtime.LockOSThread()
		if C.Py_IsInitialized() != 0 {
			ch <- errors.New("python has already been initialized by another module" +
				" but sensorbee/py needs to initialize python by itself to keep using the same main thread")
			return
		}
		C.Py_Initialize()
		if C.Py_IsInitialized() == 0 {
			ch <- errors.New("cannot initialize python")
			return
		}
		defer func() {
			C.Py_Finalize()
		}()

		// TODO: as long as mainthread uses a single goroutine, there might be no
		// need to acquire GIL because other threads never touch the interpreter.
		if C.PyEval_ThreadsInitialized() != 0 { // just in case
			ch <- errors.New("python threads are already initialized although the interpreter wasn't initialized")
			return
		}

		C.PyEval_InitThreads()                  // This call acquires the GIL.
		if C.PyEval_ThreadsInitialized() == 0 { // again, just in case
			ch <- errors.New("cannot initialize GIL")
			return
		}
		defer func() {
			C.PyEval_ReleaseThread(C.PyGILState_GetThisThreadState())
		}()

		if err := importSys(); err != nil {
			ch <- err
			return
		}
		ch <- nil
		process()
	}()

	if err := <-ch; err != nil {
		panic(err)
	}
}

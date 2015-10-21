package py

/*
#include "Python.h"
#include "datetime.h"

void init_PyDateTime() {
  PyDateTime_IMPORT;
}

int IsPyTypeDateTime(PyObject* o) {
  return PyDateTime_CheckExact(o);
}

int IsPyTypeTimeDelta(PyObject* o) {
  return PyDelta_CheckExact(o);
}

PyObject* GetPyDateTime(int year, int month, int day, int hour, int minute,
                        int second, int us) {
  return PyDateTime_FromDateAndTime(year, month, day, hour, minute, second, us);
}
*/
import "C"
import (
	"runtime"
	"time"
)

func init() {
	// Lock Native thread for initializing python
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// PyDateTime_IMPORT requires initialized python interpreter and GIL.
	tstate, err := initAndLockPython()
	if err != nil {
		panic(err)
	}
	defer C.PyEval_ReleaseThread(tstate)

	C.init_PyDateTime()
}

func IsPyTypeDateTime(o *C.PyObject) bool {
	return C.IsPyTypeDateTime(o) > 0
}

func IsPyTypeTimeDelta(o *C.PyObject) bool {
	return C.IsPyTypeTimeDelta(o) > 0
}

func getPyDateTime(t time.Time) *C.PyObject {
	us := int(t.Nanosecond() / 1e3)
	return C.GetPyDateTime(C.int(t.Year()), C.int(t.Month()), C.int(t.Day()),
		C.int(t.Hour()), C.int(t.Minute()), C.int(t.Second()), C.int(us))
}

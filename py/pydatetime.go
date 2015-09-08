package py

/*
#cgo darwin pkg-config: python-2.7
#include "Python.h"
#include "datetime.h"

void init_PyDateTime() {
  PyDateTime_IMPORT;
}

int PyDateTimeCheckExact(PyObject* o) {
  return PyDateTime_CheckExact(o);
}

PyObject* GetPyDateTime(int year, int month, int day, int hour, int minute,
                        int second, int us) {
  return PyDateTime_FromDateAndTime(year, month, day, hour, minute, second, us);
}

int PyTimeDeltaCheckExact(PyObject* o) {
  return PyDelta_CheckExact(o);
}
*/
import "C"
import (
	"time"
)

func init() {
	C.init_PyDateTime()
}

func pyDateTimeCheckExact(o *C.PyObject) bool {
	return C.PyDateTimeCheckExact(o) > 0
}

func getPyDateTime(t time.Time) *C.PyObject {
	us := int(t.Nanosecond() / 1e3)
	return C.GetPyDateTime(C.int(t.Year()), C.int(t.Month()), C.int(t.Day()),
		C.int(t.Hour()), C.int(t.Minute()), C.int(t.Second()), C.int(us))
}

func pyTimeDeltaCheckExact(o *C.PyObject) bool {
	return C.PyTimeDeltaCheckExact(o) > 0
}

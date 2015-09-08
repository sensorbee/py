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
*/
import "C"

func init() {
	C.init_PyDateTime()
}

func pyDateTimeCheckExact(o *C.PyObject) bool {
	return C.PyDateTimeCheckExact(o) > 0
}

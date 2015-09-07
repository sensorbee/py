package py

/*
#cgo darwin pkg-config: python-2.7
#include "Python.h"
*/
import "C"
import (
	"pfi/sensorbee/sensorbee/data"
	"unsafe"
)

func fromPyTypeObject(o *C.PyObject) data.Value {
	switch {
	case o.ob_type == &C.PyInt_Type:
		return data.Int(C.PyInt_AsLong(o))

	case o.ob_type == &C.PyByteArray_Type:
		bytePtr := C.PyByteArray_FromObject(o)
		charPtr := C.PyByteArray_AsString(bytePtr)
		l := C.PyByteArray_Size(o)
		return data.Blob(C.GoBytes(unsafe.Pointer(charPtr), C.int(l)))
	}
	// TODO: implement other types
	return nil
}

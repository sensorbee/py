package py

/*
#cgo pkg-config: python-2.7
#include "Python.h"
#include "datetime.h"

int IsPyTypeTrue(PyObject *o) {
  return o == Py_True;
}

int IsPyTypeFalse(PyObject *o) {
  return o == Py_False;
}

int IsPyTypeInt(PyObject *o) {
  return PyInt_CheckExact(o);
}

int IsPyTypeFloat(PyObject *o) {
  return PyFloat_CheckExact(o);
}

int IsPyTypeByteArray(PyObject *o) {
  return PyByteArray_CheckExact(o);
}

int IsPyTypeString(PyObject *o) {
  return PyString_CheckExact(o);
}

int IsPyTypeList(PyObject *o) {
  return PyList_CheckExact(o);
}

int IsPyTypeDict(PyObject *o) {
  return PyDict_CheckExact(o);
}

int IsPyTypeNone(PyObject *o) {
  return o == Py_None;
}
*/
import "C"
import (
	"pfi/sensorbee/sensorbee/data"
	"time"
	"unsafe"
)

func fromPyTypeObject(o *C.PyObject) data.Value {
	switch {
	case C.IsPyTypeTrue(o) > 0:
		return data.Bool(true)

	case C.IsPyTypeFalse(o) > 0:
		return data.Bool(false)

	case C.IsPyTypeInt(o) > 0:
		return data.Int(C.PyInt_AsLong(o))

	case C.IsPyTypeFloat(o) > 0:
		return data.Float(C.PyFloat_AsDouble(o))

	case C.IsPyTypeByteArray(o) > 0:
		bytePtr := C.PyByteArray_FromObject(o)
		charPtr := C.PyByteArray_AsString(bytePtr)
		l := C.PyByteArray_Size(o)
		return data.Blob(C.GoBytes(unsafe.Pointer(charPtr), C.int(l)))

	case C.IsPyTypeString(o) > 0:
		size := C.int(C.PyString_Size(o))
		charPtr := C.PyString_AsString(o)
		return data.String(string(C.GoBytes(unsafe.Pointer(charPtr), size)))

	case IsPyTypeDateTime(o):
		return fromTimestamp(o)

	case C.IsPyTypeList(o) > 0:
		return fromPyArray(o)

	case C.IsPyTypeDict(o) > 0:
		return fromPyMap(o)

	case C.IsPyTypeNone(o) > 0:
		return data.Null{}

	}

	return data.Null{}
}

func fromPyArray(ls *C.PyObject) data.Array {
	size := int(C.PyList_Size(ls))
	array := make(data.Array, size)
	for i := 0; i < size; i++ {
		o := C.PyList_GetItem(ls, C.Py_ssize_t(i))
		array[i] = fromPyTypeObject(o)
	}
	return array
}

func fromPyMap(o *C.PyObject) data.Map {
	m := data.Map{}

	var key, value *C.PyObject
	pos := C.Py_ssize_t(C.int(0))

	for C.int(C.PyDict_Next(o, &pos, &key, &value)) > 0 {
		// data.Map's key is only allowed string
		if key.ob_type != &C.PyString_Type {
			continue
		}
		key, _ := data.ToString(fromPyTypeObject(key))
		m[key] = fromPyTypeObject(value)
	}

	return m
}

func fromTimestamp(o *C.PyObject) data.Timestamp {
	// FIXME: this internal code should not use
	d := (*C.PyDateTime_DateTime)(unsafe.Pointer(o))
	t := time.Date(int(d.data[0])<<8|int(d.data[1]), time.Month(int(d.data[2])+1),
		int(d.data[3]), int(d.data[4]), int(d.data[5]), int(d.data[6]),
		(int(d.data[7])<<16|int(d.data[8])<<8|int(d.data[9]))*1000,
		time.UTC)

	if d.hastzinfo <= 0 {
		return data.Timestamp(t)
	}

	return fromTimestampWithTimezone(o, t)
}

// fromTimestampWithTimezone converts into data.Timestamp with UTC time zone
// from datetime with tzinfo.  All of datetime passed to Go from Python API
// must be unified into UTC time zone by this function.
//
// This function calls `utcoffset` method to acrquire offset from UTC for
// adjusting time zone.
func fromTimestampWithTimezone(o *C.PyObject, t time.Time) data.Timestamp {
	pyFunc, err := getPyFunc(o, "utcoffset")
	if err != nil {
		// Cannot get `utcoffset` function
		return data.Timestamp(t)
	}
	defer pyFunc.decRef()

	ret, err := pyFunc.CallObject(Object{})
	if ret.p == nil && err != nil {
		// Failed to execute `utcoffset` function
		return data.Timestamp(t)
	}

	if !IsPyTypeTimeDelta(ret.p) {
		// Cannot get `datetime.timedelta` instance
		return data.Timestamp(t)
	}

	// Adjust for time zone
	delta := (*C.PyDateTime_Delta)(unsafe.Pointer(ret.p))
	t = t.AddDate(0, 0, -int(delta.days))
	t = t.Add(time.Duration(-delta.seconds)*time.Second + time.Duration(-delta.microseconds)*time.Microsecond)
	return data.Timestamp(t)
}

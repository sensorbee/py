package py

/*
#cgo darwin pkg-config: python-2.7
#include "Python.h"
#include "datetime.h"
*/
import "C"
import (
	"pfi/sensorbee/sensorbee/data"
	"time"
	"unsafe"
)

func fromPyTypeObject(o *C.PyObject) data.Value {
	switch {
	// FIXME: this internal code should not use
	case o == (*C.PyObject)(unsafe.Pointer(&C._Py_TrueStruct)):
		return data.Bool(true)

	// FIXME: this internal code should not use
	case o == (*C.PyObject)(unsafe.Pointer(&C._Py_ZeroStruct)):
		return data.Bool(false)

	case o.ob_type == &C.PyInt_Type:
		return data.Int(C.PyInt_AsLong(o))

	case o.ob_type == &C.PyFloat_Type:
		return data.Float(C.PyFloat_AsDouble(o))

	case o.ob_type == &C.PyByteArray_Type:
		bytePtr := C.PyByteArray_FromObject(o)
		charPtr := C.PyByteArray_AsString(bytePtr)
		l := C.PyByteArray_Size(o)
		return data.Blob(C.GoBytes(unsafe.Pointer(charPtr), C.int(l)))

	case o.ob_type == &C.PyString_Type:
		size := C.int(C.PyString_Size(o))
		charPtr := C.PyString_AsString(o)
		return data.String(string(C.GoBytes(unsafe.Pointer(charPtr), size)))

	case pyDateTimeCheckExact(o):
		return fromTimestamp(o)

	case o.ob_type == &C.PyList_Type:
		return fromPyArray(o)

	case o.ob_type == &C.PyDict_Type:
		return fromPyMap(o)

	// FIXME: this internal code should not use
	case o == &C._Py_NoneStruct:
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
	m := ObjectModule{Object{p: o}}
	pyFunc, err := m.getPyFunc("utcoffset")
	if err != nil {
		// Cannot get `utcoffset` function
		return data.Timestamp(t)
	}
	defer pyFunc.DecRef()

	ret, err := pyFunc.CallObject(Object{})
	if ret.p == nil && err != nil {
		// Failed to execute `utcoffset` function
		return data.Timestamp(t)
	}

	if !pyTimeDeltaCheckExact(ret.p) {
		// Cannot get `datetime.timedelta` instance
		return data.Timestamp(t)
	}

	// Adjust for time zone
	delta := (*C.PyDateTime_Delta)(unsafe.Pointer(ret.p))
	t = t.AddDate(0, 0, -int(delta.days))
	t = t.Add(time.Duration(-delta.seconds)*time.Second + time.Duration(-delta.microseconds)*time.Microsecond)
	return data.Timestamp(t)
}

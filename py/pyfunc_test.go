package py

import (
	. "github.com/smartystreets/goconvey/convey"
	"pfi/sensorbee/sensorbee/data"
	"runtime"
	"testing"
)

func TestPyFunc(t *testing.T) {
	Convey("Given an initialized pyfunc test module", t, func() {
		ImportSysAndAppendPath("")

		mdl, err := LoadModule("_test_pyfunc")
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)

		obj, err := mdl.NewInstance("negator")
		So(err, ShouldBeNil)
		defer obj.DecRef()
		negator := ObjectFunc{obj.Object}

		Convey("When calling a function object properly", func() {
			arg, err := safePythonCall(func() (Object, error) {
				return convertArgsGo2Py([]data.Value{data.Int(1)})
			})
			So(err, ShouldBeNil)
			defer arg.DecRef()

			ret, err := safePythonCall(func() (Object, error) {
				return negator.callObject(arg)
			})
			Convey("it should succeed.", func() {
				So(err, ShouldBeNil)
				defer ret.DecRef()
			})
		})

		Convey("When calling a function object with incorect number of arguments", func() {
			badArg, err := safePythonCall(func() (Object, error) {
				return convertArgsGo2Py(nil)
			})
			So(err, ShouldBeNil)
			defer badArg.DecRef()

			ret2, err := safePythonCall(func() (Object, error) {
				return negator.callObject(badArg)
			})
			Convey("it should return an error.", func() {
				So(err, ShouldNotBeNil)
				So(ret2.p, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "__call__() takes exactly 2 arguments (1 given)")
			})
		})
	})
}

func safePythonCall(f func() (Object, error)) (Object, error) {
	type ret struct {
		o Object
		e error
	}

	c := make(chan ret)
	go func() {
		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)
		o, e := f()
		c <- ret{o, e}
	}()
	r := <-c
	return r.o, r.e
}

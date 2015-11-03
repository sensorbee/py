package py

import (
	. "github.com/smartystreets/goconvey/convey"
	"pfi/sensorbee/sensorbee/data"
	"runtime"
	"testing"
)

func TestGetPyFuncError(t *testing.T) {
	Convey("Given an initialized pyfunc test module", t, func() {
		ImportSysAndAppendPath("")

		mdl, err := LoadModule("_test_pyfunc")
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)
		Reset(func() {
			mdl.DecRef()
		})

		Convey("When func name is not exist", func() {
			_, err := safePythonCall(func() (Object, error) {
				o, err := getPyFunc(mdl.p, "not_exit_func")
				if err != nil {
					return Object{}, err
				}
				// this code is illegal, should not cast ObjectFunc to Object,
				// but this function is only to confirm error
				return Object{p: o.p}, nil
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "AttributeError")
		})

		Convey("When given name does not indicate function attribute", func() {
			_, err := safePythonCall(func() (Object, error) {
				o, err := getPyFunc(mdl.p, "not_func_attr")
				if err != nil {
					return Object{}, err
				}
				// this code is illegal, should not cast ObjectFunc to Object,
				// but this function is only to confirm error
				return Object{p: o.p}, nil
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not callable")
		})
	})
}

func TestPyFunc(t *testing.T) {
	Convey("Given an initialized pyfunc test module", t, func() {
		ImportSysAndAppendPath("")

		mdl, err := LoadModule("_test_pyfunc")
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)

		obj, err := mdl.NewInstance("negator", nil, nil)
		So(err, ShouldBeNil)
		Reset(func() {
			obj.DecRef()
		})
		negator := ObjectFunc{obj.Object}

		alwaysFailObj, err := mdl.NewInstance("alwaysFail", nil, nil)
		So(err, ShouldBeNil)
		Reset(func() {
			alwaysFailObj.DecRef()
		})
		alwaysFail := ObjectFunc{alwaysFailObj.Object}

		Convey("When calling a function object properly", func() {
			arg, err := safePythonCall(func() (Object, error) {
				return convertArgsGo2Py([]data.Value{data.Int(1)})
			})
			So(err, ShouldBeNil)
			Reset(func() {
				arg.DecRef()
			})

			ret, err := safePythonCall(func() (Object, error) {
				return negator.callObject(arg)
			})
			Convey("it should succeed.", func() {
				So(err, ShouldBeNil)
				Reset(func() {
					ret.DecRef()
				})
			})
		})

		Convey("When calling a function object with incorect number of arguments", func() {
			badArg, err := safePythonCall(func() (Object, error) {
				return convertArgsGo2Py(nil)
			})
			So(err, ShouldBeNil)
			Reset(func() {
				badArg.DecRef()
			})

			ret2, err := safePythonCall(func() (Object, error) {
				return negator.callObject(badArg)
			})
			Convey("it should return an error.", func() {
				So(err, ShouldNotBeNil)
				So(ret2.p, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "__call__() takes exactly 2 arguments (1 given)")
			})
		})

		Convey("When calling a function which divides by zero", func() {
			ret, err := safePythonCall(func() (Object, error) {
				return alwaysFail.callObject(Object{nil})
			})
			Convey("it should return an error with stacktrace.", func() {
				So(err, ShouldNotBeNil)
				So(ret.p, ShouldBeNil)
				So(err.Error(), ShouldStartWith, "ZeroDivisionError:")
				So(err.Error(), ShouldContainSubstring, "Traceback (most recent call last):")
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

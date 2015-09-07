package py

import (
	. "github.com/smartystreets/goconvey/convey"
	"pfi/sensorbee/sensorbee/data"
	"testing"
)

func init() {
	Initialize()
}

func TestConvertGo2PyObject(t *testing.T) {
	Convey("Given an initialized python go2py test module", t, func() {
		ImportSysAndAppendPath("")

		mdl, err := LoadModule("_test_go2py")
		defer mdl.DecRef()
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)

		Convey("When a object is full value set map", func() {
			arg := data.Map{
				"string": data.String("test"),
				"int":    data.Int(5),
				"byte":   data.Blob([]byte("ABC")),
			}
			actual, err := mdl.CallMapString("go2py", arg)
			Convey("Then function should return valid values", func() {
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "test5ABC")
			})
		})
	})
}

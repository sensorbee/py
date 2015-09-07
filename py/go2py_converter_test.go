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
				"float":  data.Float(0.1),
				"byte":   data.Blob([]byte("ABC")),
				"bool":   data.True,
				"null":   data.Null{},
				"array": data.Array{
					data.String("array-test"), data.Int(55),
				},
			}
			actual, err := mdl.Call("go2py", arg)
			Convey("Then function should return valid values", func() {
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "test_5_0.1_ABC_True_None_array-test_55")
			})
		})
	})
}

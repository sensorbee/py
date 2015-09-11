package py

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"pfi/sensorbee/sensorbee/data"
	"testing"
	"time"
)

func init() {
	Initialize()
}

func TestConvertGo2PyObject(t *testing.T) {
	Convey("Given an initialized python go2py test module", t, func() {
		ImportSysAndAppendPath("")

		mdl, err := LoadModule("_test_go2py")
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)

		type argAndExpected struct {
			arg      data.Value
			expected string
		}

		Convey("When set an object", func() {
			now := time.Now()
			nowStr := now.Format("2006-01-02 15:04:05.999999")
			values := map[string]argAndExpected{
				"string": argAndExpected{data.String("test"), "test"},
				"int":    argAndExpected{data.Int(9), "9"},
				"float":  argAndExpected{data.Float(0.9), "0.9"},
				"byte":   argAndExpected{data.Blob([]byte("ABC")), "ABC"},
				"true":   argAndExpected{data.True, "True"},
				"false":  argAndExpected{data.False, "False"},
				"time":   argAndExpected{data.Timestamp(now), nowStr},
				"null":   argAndExpected{data.Null{}, "None"},
			}
			for k, v := range values {
				v := v
				msg := fmt.Sprintf("Then function should return string value: %v", k)
				Convey(msg, func() {
					actual, err := mdl.Call("go2py_tostr", v.arg)
					So(err, ShouldBeNil)
					So(actual, ShouldEqual, v.expected)
				})
			}
		})

		Convey("When set map in map and map in array", func() {
			arg := data.Map{
				"string": data.String("test"),
				"map": data.Map{
					"instr": data.String("test2"),
				},
				"array": data.Array{
					data.String("array-test"), data.Int(55),
				},
			}
			actual, err := mdl.Call("go2py_mapinmap", arg)
			Convey("Then function should return valid values", func() {
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "test_test2_array-test_55")
			})
		})

		Convey("When set array in array and map", func() {
			arg := data.Array{
				data.Array{
					data.String("test"), data.Int(55),
				},
				data.Map{
					"map": data.String("inmap"),
				},
			}
			actual, err := mdl.Call("go2py_arrayinmap", arg)
			Convey("Then function should return valid values", func() {
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "test_55_inmap")
			})
		})

		Reset(func() {
			mdl.DecRef()
		})
	})
}

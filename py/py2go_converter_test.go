package py

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"pfi/sensorbee/sensorbee/data"
	"testing"
)

func init() {
	Initialize()
}

func TestConvertPyObject2Go(t *testing.T) {
	Convey("Given an initialized python py2go test module", t, func() {
		ImportSysAndAppendPath("")

		mdl, err := LoadModule("_test_py2go")
		defer mdl.DecRef()
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)

		returnTypes := []struct {
			typeName string
			expected data.Value
		}{
			{"true", data.Bool(true)},
			{"false", data.Bool(false)},
			{"int", data.Int(123)},
			{"float", data.Float(1.0)},
			{"string", data.String("ABC")},
			{"bytearray", data.Blob([]byte("abcdefg"))},
			{"map", data.Map{"key1": data.Int(123), "key2": data.String("str")}},
			{"nested_map", data.Map{"key1": data.Map{"key2": data.Int(123)}}},
			{"array", data.Array{data.Int(1), data.Int(2), data.Map{"key": data.Int(3)}}},
			{"none", data.Null{}},
		}

		for _, r := range returnTypes {
			r := r
			Convey(fmt.Sprint("When calling a function that returns ", r.typeName), func() {
				actual, err := mdl.Call(fmt.Sprintf("return_%s", r.typeName))
				So(err, ShouldBeNil)

				Convey("Then the function should return valid value", func() {
					So(actual, ShouldResemble, r.expected)
				})
			})
		}
	})
}

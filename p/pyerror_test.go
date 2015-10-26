package py

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPyError(t *testing.T) {
	Convey("When importing exceptions module and extract SyntaxError type", t, func() {
		mdl, err := LoadModule("exceptions")
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)

		syntaxError, err := mdl.GetClass("SyntaxError")
		So(err, ShouldBeNil)
		Reset(func() {
			syntaxError.DecRef()
		})
		Convey("applying isSyntaxError to SyntaxError should return true.", func() {
			So(isSyntaxError(syntaxError.p), ShouldBeTrue)
		})

		environmentError, err := mdl.GetClass("EnvironmentError")
		So(err, ShouldBeNil)
		Reset(func() {
			environmentError.DecRef()
		})
		Convey("applying isSyntaxError to EnvironmentError should return false.", func() {
			So(isSyntaxError(environmentError.p), ShouldBeFalse)
		})
	})
}

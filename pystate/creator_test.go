package pystate

import (
	. "github.com/smartystreets/goconvey/convey"
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
	"testing"
)

func TestCreateState(t *testing.T) {
	cc := &core.ContextConfig{}
	ctx := core.NewContext(cc)
	Convey("Given a pystate creator", t, func() {
		ct := Creator{}
		Convey("When the parameter has all required values", func() {
			params := data.Map{
				"module_name": data.String("_test_creator_module"),
				"class_name":  data.String("TestClass"),
			}
			Convey("Then the state should be created and set default value", func() {
				state, err := ct.CreateState(ctx, params)
				So(err, ShouldBeNil)
				Reset(func() {
					state.Terminate(ctx)
				})
				ps, ok := state.(*pyState)
				So(ok, ShouldBeTrue)
				So(ps.modulePath, ShouldEqual, "")
				So(ps.moduleName, ShouldEqual, "_test_creator_module")

				ctx.SharedStates.Add("creator_test", "creator_test", state)
				Reset(func() {
					ctx.SharedStates.Remove("creator_test")
				})
				Convey("When prepare to be called one instance method", func() {
					Convey("Then exist instance method should be called", func() {
						dt := data.String("test")
						v, err := Func(ctx, "creator_test", "write", dt)
						So(err, ShouldBeNil)
						So(v, ShouldEqual, `called! arg is "test"`)
					})
					Convey("Then not exist instance method should not be called and return error", func() {
						_, err = Func(ctx, "creator_test", "not_exist_method")
						So(err, ShouldNotBeNil)
					})
				})
			})
		})

		Convey("When the parameter has constructor arguments", func() {
			params := data.Map{
				"module_name": data.String("_test_creator_module"),
				"class_name":  data.String("TestClass2"),
				"v1":          data.String("init_test"),
				"v2":          data.String("init_test2"),
			}
			Convey("Then the state should be created with constructor arguments", func() {
				state, err := ct.CreateState(ctx, params)
				So(err, ShouldBeNil)
				Reset(func() {
					state.Terminate(ctx)
				})

				ctx.SharedStates.Add("creator_test2", "creator_test2", state)
				Reset(func() {
					ctx.SharedStates.Remove("creator_test2")
				})
				v, err := Func(ctx, "creator_test2", "confirm")
				So(err, ShouldBeNil)
				So(v, ShouldEqual, `constructor init arg is v1=init_test, v2=init_test2`)
			})
		})

		SkipConvey("When the parameter has constructor arguments which lack optional value", func() {
			params := data.Map{
				"module_name": data.String("_test_creator_module"),
				"class_name":  data.String("TestClass3"),
				"a":           data.Int(55),
			}
			Convey("Then the state should be created and initialized with only required value", func() {
				state, err := ct.CreateState(ctx, params)
				So(err, ShouldBeNil)
				Reset(func() {
					state.Terminate(ctx)
				})

				ctx.SharedStates.Add("creator_test3", "creator_test3", state)
				Reset(func() {
					ctx.SharedStates.Remove("creator_test3")
				})
				v, err := Func(ctx, "creator_test3", "confirm")
				So(err, ShouldBeNil)
				So(v, ShouldEqual, "constructor init arg is a=55, b=b, c={}")
			})
		})

		Convey("When the parameter has required value and option value", func() {
			params := data.Map{
				"module_name":  data.String("_test_creator_module"),
				"class_name":   data.String("TestClass"),
				"write_method": data.String("write"),
			}
			Convey("Then the state should be created and it is writable", func() {
				state, err := ct.CreateState(ctx, params)
				So(err, ShouldBeNil)
				Reset(func() {
					state.Terminate(ctx)
				})
				ps, ok := state.(*pyWritableState)
				So(ok, ShouldBeTrue)
				So(ps.writeFuncName, ShouldEqual, "write")

				t := &core.Tuple{}
				err = ps.Write(ctx, t)
				So(err, ShouldBeNil)
			})
		})

		Convey("When the parameter lacks module name", func() {
			params := data.Map{
				"class_name": data.String("TestClass"),
			}
			Convey("Then a state should not be created", func() {
				state, err := ct.CreateState(ctx, params)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "module_name")
				So(state, ShouldBeNil)
			})
		})

		Convey("When the parameter lacks class name", func() {
			params := data.Map{
				"module_name": data.String("_test_creator_module"),
			}
			Convey("Then a state should not be created", func() {
				state, err := ct.CreateState(ctx, params)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "class_name")
				So(state, ShouldBeNil)
			})
		})
	})
}

package py

import (
	. "github.com/smartystreets/goconvey/convey"
	"pfi/sensorbee/py/mainthread"
	"pfi/sensorbee/sensorbee/data"
	"testing"
)

func TestLoadModuleError(t *testing.T) {
	Convey("Given a python interpreter set up default import path", t, func() {
		mainthread.AppendSysPath("")

		Convey("When get module with not exist module name", func() {
			_, err := LoadModule("not_exist_module")
			Convey("Then an error should be occurred", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "ImportError")
			})
		})

		Convey("When get syntax error module", func() {
			_, err := LoadModule("_test_syntax_error")
			Convey("Then an error should be occurred", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "SyntaxError")
			})
		})
	})
}

func TestNewInstanceAndStateness(t *testing.T) {
	Convey("Given an initialized python module", t, func() {

		mainthread.AppendSysPath("")

		mdl, err := LoadModule("_test_new_instance")
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)

		Convey("When get an invalid class instance", func() {
			_, err := mdl.NewInstance("NonexistentClass", nil, nil)
			Convey("Then an error should be occurred", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "AttributeError")
			})
		})

		Convey("When get a new test python instance", func() {
			ins, err := mdl.NewInstance("PythonTest", nil, nil)
			Convey("Then process should get a python instance from the module", func() {
				So(err, ShouldBeNil)
				So(ins, ShouldNotBeNil)
				Reset(func() {
					ins.Release()
				})

				Convey("And when call a logger function", func() {
					actual, err := ins.Call("logger", data.String("test"))
					Convey("Then process should get a string", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, "initialized_test")
					})

					Convey("And when call a logger function again", func() {
						actual2, err := ins.Call("logger", data.String("test"))
						Convey("Then process should get a string", func() {
							So(err, ShouldBeNil)
							So(actual2, ShouldEqual, "initialized_test_test")
						})

						Convey("And when get a new test python instance", func() {
							ins2, err := mdl.NewInstance("PythonTest", nil, nil)
							Convey("Then process should get another instance from the module", func() {
								So(err, ShouldBeNil)
								So(ins2, ShouldNotBeNil)
								Reset(func() {
									ins2.Release()
								})

								Convey("And when call a logger function", func() {
									actual3, err1 := ins.Call("logger", data.String("t"))
									actual4, err2 := ins2.Call("logger", data.String("t"))
									Convey("Then process should get a string", func() {
										So(err1, ShouldBeNil)
										So(err2, ShouldBeNil)
										So(actual3, ShouldEqual, "initialized_test_test_t")
										So(actual4, ShouldEqual, "initialized_t")
									})
								})
							})
						})
					})
				})
			})
		})

		Convey("When get a new test python instance with param", func() {
			params := data.String("python_test")
			ins, err := mdl.NewInstance("PythonTest2", []data.Value{params}, nil)

			Convey("Then process should get instance and set values", func() {
				So(err, ShouldBeNil)
				So(ins, ShouldNotBeNil)
				Reset(func() {
					ins.Release()
				})

				actual, err := ins.Call("get_a")
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "python_test")

				Convey("And when get another python instance with param", func() {
					params2 := data.String("python_test2")
					ins2, err := mdl.NewInstance("PythonTest2", []data.Value{params2}, nil)

					Convey("Then process should get another instance and set values", func() {
						So(err, ShouldBeNil)
						So(ins2, ShouldNotBeNil)
						Reset(func() {
							ins2.Release()
						})

						actual2, err := ins2.Call("get_a")
						So(err, ShouldBeNil)
						So(actual2, ShouldEqual, "python_test2")

						react, err := ins.Call("get_a")
						So(err, ShouldBeNil)
						So(react, ShouldEqual, "python_test")

					})
				})
			})
		})

		Convey("When get a new test python instance with invalid param", func() {
			params := data.String("unnecessary_param")
			_, err := mdl.NewInstance("PythonTest", []data.Value{params}, nil)
			Convey("Then an error should be occurred", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "TypeError")
			})
		})

		Convey("When get a new class instance", func() {
			class, err := mdl.GetClass("PythonTest3")
			Convey("Then process should get PythonTest3 class", func() {
				So(err, ShouldBeNil)
			})
			Reset(func() {
				class.Release()
			})

			Convey("Then process should get a string", func() {
				actual, err := class.Call("get_static_value")
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "class_value")
			})

			Convey("Then process should get a new instance", func() {
				obj, err := class.CallDirect("get_instance", nil, nil)
				So(err, ShouldBeNil)
				ins := &ObjectInstance{obj}
				actual, err := ins.Call("get_instance_str")
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "instance method test1")
			})

			Convey("Then process should get a new instance with args", func() {
				args := []data.Value{data.Int(55)}
				kwdArgs := data.Map{
					"v1": data.String("homhom"),
				}
				obj, err := class.CallDirect("get_instance2", args, kwdArgs)
				So(err, ShouldBeNil)
				ins := &ObjectInstance{obj}
				actual, err := ins.Call("confirm")
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "55_5_{'v1': 'homhom'}")
			})
		})

		Convey("When get a new child class of PythonTest3 instance", func() {
			class, err := mdl.GetClass("ChildClass")
			Convey("Then process should get ChildClass class", func() {
				So(err, ShouldBeNil)
			})
			Reset(func() {
				class.Release()
			})

			Convey("Then process should get a string", func() {
				actual, err := class.Call("get_class_value")
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "instance_value")
			})
		})

		Convey("When get an invalid class instance by GetClass", func() {
			_, err := mdl.GetClass("NonexistentClass")
			Convey("Then an error should be occurred", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "AttributeError")
			})
		})
	})
}

func TestNewInstanceWithKeywordArgument(t *testing.T) {
	Convey("Given an initialized python module", t, func() {

		mainthread.AppendSysPath("")

		mdl, err := LoadModule("_test_new_instance")
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)
		Reset(func() {
			mdl.Release()
		})

		Convey("When get a new test python instance with empty map", func() {
			ins, err := mdl.NewInstance("PythonTest", nil, data.Map{})
			Reset(func() {
				ins.Release()
			})

			Convey("Then process should get a python instance from the module", func() {
				So(err, ShouldBeNil)
				So(ins, ShouldNotBeNil)
			})
		})

		Convey("When get a new test python instance with named arguments", func() {
			arg := data.Map{
				"a": data.Int(1),
				"b": data.Int(2),
				"c": data.Int(3),
				"d": data.Int(4),
			}
			ins, err := mdl.NewInstance("PythonTestForKwd", nil, arg)
			Reset(func() {
				ins.Release()
			})

			Convey("Then process should get instance and set values", func() {
				So(err, ShouldBeNil)
				So(ins, ShouldNotBeNil)

				actual, err := ins.Call("confirm_init")
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "1_2_{'c': 3, 'd': 4}")
			})
		})

		Convey("When get a new test python instance with named arguments which lacks optional value", func() {
			arg := data.Map{
				"a": data.Int(1),
			}
			ins, err := mdl.NewInstance("PythonTestForKwd", nil, arg)
			Reset(func() {
				ins.Release()
			})

			Convey("Then process should get instance and set values", func() {
				So(err, ShouldBeNil)
				So(ins, ShouldNotBeNil)

				actual, err := ins.Call("confirm_init")
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "1_5_{}")
			})
		})

		Convey("When get a new test python instance with named and non-named arguments mixed", func() {
			arg := data.Int(1)
			kwdArg := data.Map{
				"v1": data.String("homhom"),
			}
			ins, err := mdl.NewInstance("PythonTestForKwd", []data.Value{arg}, kwdArg)
			Reset(func() {
				ins.Release()
			})

			Convey("Then process should get instance and set values", func() {
				So(err, ShouldBeNil)
				So(ins, ShouldNotBeNil)

				actual, err := ins.Call("confirm_init")
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "1_5_{'v1': 'homhom'}")
			})
		})

		Convey("When constructor arguments lacks required value", func() {
			_, err := mdl.NewInstance("PythonTestForKwd", nil, data.Map{})

			Convey("Then an error should be get", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "TypeError")
			})
		})
	})
}

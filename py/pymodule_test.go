package py

import (
	. "github.com/smartystreets/goconvey/convey"
	"pfi/sensorbee/sensorbee/data"
	"testing"
)

func TestNewInstanceAndStateness(t *testing.T) {
	Convey("Given an initialized python module", t, func() {

		ImportSysAndAppendPath("")

		mdl, err := LoadModule("_test_new_instance")
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)

		Convey("When get an invalid class instance", func() {
			_, err := mdl.NewInstance("NonexistentClass")
			Convey("Then an error should be occurred", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When get a new test python instance", func() {
			ins, err := mdl.NewInstance("PythonTest")
			Convey("Then process should get PyModule", func() {
				So(err, ShouldBeNil)
				So(ins, ShouldNotBeNil)
				Reset(func() {
					ins.DecRef()
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
							ins2, err := mdl.NewInstance("PythonTest")
							Convey("Then process should get PyModule", func() {
								So(err, ShouldBeNil)
								So(ins2, ShouldNotBeNil)
								Reset(func() {
									ins2.DecRef()
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
			ins, err := mdl.NewInstance("PythonTest2", params)

			Convey("Then process should get instance and set values", func() {
				So(err, ShouldBeNil)
				So(ins, ShouldNotBeNil)
				Reset(func() {
					ins.DecRef()
				})

				actual, err := ins.Call("get_a")
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, "python_test")

				Convey("And when get another python instance with param", func() {
					params2 := data.String("python_test2")
					ins2, err := mdl.NewInstance("PythonTest2", params2)

					Convey("Then process should get another instance and set values", func() {
						So(err, ShouldBeNil)
						So(ins2, ShouldNotBeNil)
						Reset(func() {
							ins2.DecRef()
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
			_, err := mdl.NewInstance("PythonTest", params)
			Convey("Then an error should be occurred", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

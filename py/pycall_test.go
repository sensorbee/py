package py

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/ugorji/go/codec"
	"testing"
)

func init() {
	// goconvey is call same function several times, so in order to call
	// `Initialize` only once, use `init`
	Initialize()
}

func TestPythonCall(t *testing.T) {
	Convey("Given an initialized python interpreter", t, func() {

		ImportSysAndAppendPath("")

		Convey("When get invalid module", func() {
			_, err := LoadModule("notexist")
			Convey("Then an error should be occurred", func() {
				So(err, ShouldNotBeNil)
			})
		})
		Convey("When get valid module", func() {
			mdl, err := LoadModule("_test")
			defer mdl.DecRef()
			Convey("Then process should get PyModule", func() {
				So(err, ShouldBeNil)
				So(mdl, ShouldNotBeNil)

				Convey("And when call invalid function", func() {
					_, err := mdl.CallIntInt("notFoundMethod", 1)
					Convey("Then an error should be occurred", func() {
						So(err, ShouldNotBeNil)
					})
				})

				Convey("And when call int-int function", func() {
					actual, err := mdl.CallIntInt("tenTimes", 3)
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, 30)
					})
				})

				Convey("And when call none-string function", func() {
					actual, err := mdl.CallNoneString("logger")
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, "called")
					})
				})

				Convey("And when call none-2string function", func() {
					actual1, actual2, err := mdl.CallNone2String("twoLogger")
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(actual1, ShouldEqual, "called1")
						So(actual2, ShouldEqual, "called2")
					})
				})

				Convey("And when call string-string function", func() {
					actual, err := mdl.CallStringString("plusSuffix", "test")
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, "test_through_python")
					})
				})

				Convey("And when call specialized function", func() {
					h := &codec.MsgpackHandle{}
					var out []byte
					enc := codec.NewEncoderBytes(&out, h)

					data := []float32{1.1, 1.2, 1.3, 1.4, 1.5}
					data2 := []int{9, 8, 7., 6, 5}
					ds := map[string]interface{}{}
					ds["data"] = data
					ds["target"] = data2
					ds["model"] = []byte{}
					err := enc.Encode(ds)
					So(err, ShouldBeNil)
					So(len(out), ShouldNotEqual, 0)
					actual1, err := mdl.CallByteByte("loadMsgPack", out)
					Convey("Then function should return valid result", func() {
						So(err, ShouldBeNil)
						So(len(actual1), ShouldNotEqual, 0)

						var ds2 map[string]interface{}
						dec := codec.NewDecoderBytes(actual1, h)
						dec.Decode(&ds2)
						model, ok := ds2["model"]
						So(ok, ShouldBeTrue)
						modelByte, ok := model.([]byte)
						So(ok, ShouldBeTrue)
						So(len(modelByte), ShouldNotEqual, 0)

						log, ok := ds2["log"]
						So(ok, ShouldBeTrue)
						logByte, ok := log.([]byte)
						So(ok, ShouldBeTrue)
						So(string(logByte), ShouldEqual, "done")

						Convey("And when pass pickle data", func() {
							var out2 []byte
							enc2 := codec.NewEncoderBytes(&out2, h)
							err = enc2.Encode(ds2)
							So(err, ShouldBeNil)
							actual2, err := mdl.CallByteByte("loadMsgPack", out2)

							Convey("Then function should return model again", func() {
								So(err, ShouldBeNil)
								So(len(actual2), ShouldNotEqual, 0)
								So(string(actual2), ShouldResemble,
									"\x82\xa5model\xae\x80\x02U\aTEST_req\x00.\xa3log\xa4done")

							})
						})
					})
				})
			})
		})
	})
}

func TestPythonInstanceCall(t *testing.T) {
	Convey("Given an initialized python module", t, func() {

		ImportSysAndAppendPath("")

		mdl, err := LoadModule("_instance_test")
		So(err, ShouldBeNil)
		So(mdl, ShouldNotBeNil)

		Convey("When get a new test python instance", func() {
			ins, err := mdl.GetInstance("PythonTest")
			defer ins.DecRef()
			Convey("Then process should get PyModule", func() {
				So(err, ShouldBeNil)
				So(ins, ShouldNotBeNil)

				Convey("And when call a logger function", func() {
					actual, err := ins.CallStringString("logger", "test")
					Convey("Then process should get a string", func() {
						So(err, ShouldBeNil)
						So(actual, ShouldEqual, "initialized_test")
					})

					Convey("And when call a logger function again", func() {
						actual2, err := ins.CallStringString("logger", "test")
						Convey("Then process should get a string", func() {
							So(err, ShouldBeNil)
							So(actual2, ShouldEqual, "initialized_test_test")
						})

						Convey("And when get a new test python instance", func() {
							ins2, err := mdl.GetInstance("PythonTest")
							defer ins2.DecRef()
							Convey("Then process should get PyModule", func() {
								So(err, ShouldBeNil)
								So(ins2, ShouldNotBeNil)

								Convey("And when call a logger function", func() {
									actual3, err1 := ins.CallStringString("logger", "t")
									actual4, err2 := ins2.CallStringString("logger", "t")
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
	})
}

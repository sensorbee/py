package plugin

import (
	"gopkg.in/sensorbee/py.v0/pystate"
	"gopkg.in/sensorbee/sensorbee.v0/bql/udf"
)

func init() {
	udf.MustRegisterGlobalUDSCreator("pystate", &pystate.Creator{})
	udf.MustRegisterGlobalUDF("pystate_func", udf.MustConvertGeneric(pystate.CallMethod))
}

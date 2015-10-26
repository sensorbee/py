package plugin

import (
	"pfi/sensorbee/py/pystate"
	"pfi/sensorbee/sensorbee/bql/udf"
)

func init() {
	udf.MustRegisterGlobalUDSCreator("pystate", &pystate.Creator{})
	udf.MustRegisterGlobalUDF("pystate_func", udf.MustConvertGeneric(pystate.Func))
}

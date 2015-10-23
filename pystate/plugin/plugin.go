package plugin

import (
	"pfi/sensorbee/py/pystate"
	"pfi/sensorbee/sensorbee/bql/udf"
)

func init() {
	udf.MustRegisterGlobalUDSCreator("pystate", &pystate.PyStateCreator{})

	udf.MustRegisterGlobalUDF("pystate_func",
		udf.MustConvertGeneric(pystate.PyMLPredict))
}

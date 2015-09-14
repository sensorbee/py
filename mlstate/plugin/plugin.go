package plugin

import (
	"pfi/sensorbee/pystate/mlstate"
	"pfi/sensorbee/sensorbee/bql/udf"
)

func init() {
	udf.MustRegisterGlobalUDSCreator("pymlstate", &mlstate.PyMLStateCreator{})

	udf.MustRegisterGlobalUDF("pymlstate_fit",
		udf.MustConvertGeneric(mlstate.PyMLFit))
	udf.MustRegisterGlobalUDF("pymlstate_predict",
		udf.MustConvertGeneric(mlstate.PyMLPredict))
}

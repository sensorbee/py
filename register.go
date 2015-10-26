package py

import (
	"pfi/sensorbee/py/pystate"
	"pfi/sensorbee/sensorbee/bql/udf"
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
)

type defaultCreator struct {
	modulePath      string
	moduleName      string
	className       string
	writeMethodName string
}

func (c *defaultCreator) CreateState(ctx *core.Context, params data.Map) (
	core.SharedState, error) {
	return pystate.New(c.modulePath, c.moduleName, c.className, c.writeMethodName,
		params)
}

// MustRegisterPythonUDSCreator is like MustRegisterGlobalUDSCreator for Python
// instance.
func MustRegisterPythonUDSCreator(typeName string, modulePath string,
	moduleName string, className string, writeMethodName string) {
	udf.MustRegisterGlobalUDSCreator(typeName, &defaultCreator{
		modulePath:      modulePath,
		moduleName:      moduleName,
		className:       className,
		writeMethodName: writeMethodName,
	})
}

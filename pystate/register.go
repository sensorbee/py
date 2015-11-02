package py

import (
	"fmt"
	py "pfi/sensorbee/py/p"
	"pfi/sensorbee/py/pystate"
	"pfi/sensorbee/sensorbee/bql/udf"
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
	"sync"
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

// MustRegisterPyUDSCreator is like MustRegisterGlobalUDSCreator for Python
// instance.
func MustRegisterPyUDSCreator(typeName string, modulePath string,
	moduleName string, className string, writeMethodName string) {
	udf.MustRegisterGlobalUDSCreator(typeName, &defaultCreator{
		modulePath:      modulePath,
		moduleName:      moduleName,
		className:       className,
		writeMethodName: writeMethodName,
	})
}

type defaultFunc struct {
	mdl      py.ObjectModule
	funcName string

	mu sync.RWMutex
}

func (f *defaultFunc) Func(ctx *core.Context, args ...data.Value) (data.Value,
	error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.mdl.Call(f.funcName, args...)
}

func MustRegisterPyUDF(udfName string, modulePath string, moduleName string,
	funcName string) {
	py.ImportSysAndAppendPath(modulePath)
	mdl, err := py.LoadModule(moduleName)
	if err != nil {
		panic(fmt.Errorf("py.MustRegisterPyUDF: cannot register '%v': %v",
			udfName, err))
	}
	df := defaultFunc{
		mdl:      mdl,
		funcName: funcName,
	}
	udf.MustRegisterGlobalUDF(udfName, udf.MustConvertGeneric(df.Func))
}

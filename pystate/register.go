package pystate

import (
	"fmt"
	"pfi/sensorbee/py"
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
	return New(c.modulePath, c.moduleName, c.className, c.writeMethodName,
		params)
}

// MustRegisterPyUDSCreator is like MustRegisterGlobalUDSCreator for Python
// instance, just an alias of pystate.
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

// MustRegisterPyUDF is like MustRegisterGlobalUDF for Python module method.
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

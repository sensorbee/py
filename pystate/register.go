package pystate

import (
	"fmt"
	"gopkg.in/sensorbee/py.v0"
	"gopkg.in/sensorbee/py.v0/mainthread"
	"gopkg.in/sensorbee/sensorbee.v0/bql/udf"
	"gopkg.in/sensorbee/sensorbee.v0/core"
	"gopkg.in/sensorbee/sensorbee.v0/data"
	"sync"
)

type defaultCreator struct {
	baseParams *BaseParams
}

func (c *defaultCreator) CreateState(ctx *core.Context, params data.Map) (
	core.SharedState, error) {
	return New(c.baseParams, params)
}

// MustRegisterPyUDSCreator is like MustRegisterGlobalUDSCreator for Python
// instance, just an alias of pystate.
func MustRegisterPyUDSCreator(typeName string, modulePath string,
	moduleName string, className string, writeMethodName string) {
	udf.MustRegisterGlobalUDSCreator(typeName, &defaultCreator{
		baseParams: &BaseParams{
			ModulePath:      modulePath,
			ModuleName:      moduleName,
			ClassName:       className,
			WriteMethodName: writeMethodName,
		},
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
	mainthread.AppendSysPath(modulePath)
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

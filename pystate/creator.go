package pystate

import (
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
)

var (
	modulePath      = data.MustCompilePath("module_path")
	moduleNamePath  = data.MustCompilePath("module_name")
	classNamePath   = data.MustCompilePath("class_name")
	writeMethodPath = data.MustCompilePath("write_method")
)

// PyStateCreator is used by BQL to create or load Multiple Layer Classification
// State as a UDS.
type PyStateCreator struct {
}

// CreateState creates `core.SharedState`
//
// * module_path:  Directory path of python module path, default is ''.
// * module_name:  Python module name, required.
// * class_name:   Python class name, required.
// * write_method: [TODO]
//
// other rest parameters will set python constructor arguments, the arguments'
// type is dictionary.
func (c *PyStateCreator) CreateState(ctx *core.Context, params data.Map) (
	core.SharedState, error) {
	var err error
	mdlPathName := ""
	if mp, err := params.Get(modulePath); err == nil {
		if mdlPathName, err = data.AsString(mp); err != nil {
			return nil, err
		}
		delete(params, "module_path")
	}
	mn, err := params.Get(moduleNamePath)
	if err != nil {
		return nil, err
	}
	delete(params, "module_name")
	moduleName, err := data.AsString(mn)
	if err != nil {
		return nil, err
	}

	cn, err := params.Get(classNamePath)
	if err != nil {
		return nil, err
	}
	delete(params, "class_name")
	className, err := data.AsString(cn)
	if err != nil {
		return nil, err
	}

	return NewPyState(mdlPathName, moduleName, className, params)
}

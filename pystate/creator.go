package pystate

import (
	"io"
	"pfi/sensorbee/sensorbee/bql/udf"
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
)

var (
	modulePath      = data.MustCompilePath("module_path")
	moduleNamePath  = data.MustCompilePath("module_name")
	classNamePath   = data.MustCompilePath("class_name")
	writeMethodPath = data.MustCompilePath("write_method")
)

// Creator is used by BQL to create state as a UDS.
type Creator struct {
}

var _ udf.UDSLoader = &Creator{}

// CreateState creates `core.SharedState`
//
// * module_path:  Directory path of python module path, default is ''.
// * module_name:  Python module name, required.
// * class_name:   Python class name, required.
// * write_method: [optional] Python method name when call write in.
//
// other rest parameters will set python constructor arguments, the arguments
// are set for named arguments.
func (c *Creator) CreateState(ctx *core.Context, params data.Map) (
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

	writeMethodName := ""
	if wmn, err := params.Get(writeMethodPath); err == nil {
		if writeMethodName, err = data.AsString(wmn); err != nil {
			return nil, err
		}
		delete(params, "write_method")
	}

	return New(mdlPathName, moduleName, className, writeMethodName, params)
}

// LoadState loads saved state.
func (c *Creator) LoadState(ctx *core.Context, r io.Reader, params data.Map) (
	core.SharedState, error) {
	s := pyState{}
	if err := s.Load(ctx, r, params); err != nil {
		return nil, err
	}

	if s.writeMethodName != "" {
		return &pyWritableState{
			s,
		}, nil
	}
	return &s, nil
}

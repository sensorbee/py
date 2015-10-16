package mlstate

import (
	"fmt"
	"io"
	"pfi/sensorbee/sensorbee/bql/udf"
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
)

var (
	modulePath         = data.MustCompilePath("module_path")
	moduleNamePath     = data.MustCompilePath("module_name")
	classNamePath      = data.MustCompilePath("class_name")
	batchTrainSizePath = data.MustCompilePath("batch_train_size")
)

// PyMLStateCreator is used by BQL to create or load Multiple Layer Classification
// State as a UDS.
type PyMLStateCreator struct {
}

var _ udf.UDSLoader = &PyMLStateCreator{}

// CreateState creates `core.SharedState`
//
// * module_path:      Directory path of python module path, default is ''.
// * module_name:      Python module name, required.
// * class_name:       Python class name, required.
// * batch_train_size: Batch size of training. Created state is SharedSink, when
//                     calling "fit" function, the state send train data as
//                     array, which length is "batch_train_size". Default is 10.
func (c *PyMLStateCreator) CreateState(ctx *core.Context, params data.Map) (
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

	batchSize := 10
	if bs, err := params.Get(batchTrainSizePath); err == nil {
		var batchSize64 int64
		if batchSize64, err = data.AsInt(bs); err != nil {
			return nil, err
		}
		if batchSize64 <= 0 {
			return nil, fmt.Errorf("batch_train_size must be greater than 0")
		}
		batchSize = int(batchSize64)
		delete(params, "batch_train_size")
	}

	return NewPyMLState(mdlPathName, moduleName, className, batchSize, params)
}

// LoadState is same as CREATE STATE.
func (c *PyMLStateCreator) LoadState(ctx *core.Context, r io.Reader, params data.Map) (
	core.SharedState, error) {
	return c.CreateState(ctx, params)
}

package mlstate

import (
	"fmt"
	"io"
	"pfi/sensorbee/sensorbee/bql/udf"
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
)

var (
	modulePath         = "module_path"
	moduleNamePath     = "module_name"
	classNamePath      = "class_name"
	batchTrainSizePath = "batch_train_size"
	modelFilePath      = "model_file_path"
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
// * model_file_path:  A model to use in fit and predict. The model style is
//                     depend on Python implementation. Default is ''.
func (c *PyMLStateCreator) CreateState(ctx *core.Context, params data.Map) (
	core.SharedState, error) {
	var err error
	mdlPathName := ""
	if mp, ok := params[modulePath]; ok {
		if mdlPathName, err = data.AsString(mp); err != nil {
			return nil, err
		}
	}
	mn, ok := params[moduleNamePath]
	if !ok {
		return nil, fmt.Errorf("module_name is not specified")
	}
	moduleName, err := data.AsString(mn)
	if err != nil {
		return nil, err
	}

	cn, ok := params[classNamePath]
	if !ok {
		return nil, fmt.Errorf("class_name is not specified")
	}
	className, err := data.AsString(cn)
	if err != nil {
		return nil, err
	}

	batchSize := 10
	if bs, ok := params[batchTrainSizePath]; ok {
		var batchSize64 int64
		if batchSize64, err = data.AsInt(bs); err != nil {
			return nil, err
		}
		if batchSize64 <= 0 {
			return nil, fmt.Errorf("batch_train_size must be greater than 0")
		}
		batchSize = int(batchSize64)
	}

	modelPath := ""
	if mp, ok := params[modelFilePath]; ok {
		if modelPath, err = data.AsString(mp); err != nil {
			return nil, err
		}
	}

	return NewPyMLState(mdlPathName, moduleName, className, batchSize, modelPath)
}

// LoadState is same as CREATE STATE, but "model_file_path" is required.
func (c *PyMLStateCreator) LoadState(ctx *core.Context, r io.Reader, params data.Map) (
	core.SharedState, error) {
	if _, ok := params[modelFilePath]; !ok {
		return nil, fmt.Errorf("model_file_path is not specified")
	}

	return c.CreateState(ctx, params)
}

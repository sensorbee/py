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
)

// PyMLStateCreator is used by BQL to create or load AROWState as a UDS.
type PyMLStateCreator struct {
}

var _ udf.UDSLoader = &PyMLStateCreator{}

func (c *PyMLStateCreator) CreateState(ctx *core.Context, params data.Map) (core.SharedState, error) {
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

	return NewPyMLState(mdlPathName, moduleName, className, batchSize)
}

func (c *PyMLStateCreator) LoadState(ctx *core.Context, r io.Reader, params data.Map) (core.SharedState, error) {
	return nil, fmt.Errorf("pymlstate doesn't support LoadState")
}

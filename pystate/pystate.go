package pystate

import (
	"fmt"
	"pfi/sensorbee/pystate/py"
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
)

// PyState is python instance specialized to multiple layer classification.
type PyState struct {
	modulePath string
	moduleName string
	className  string

	ins py.ObjectInstance
}

// NewPyState creates `core.SharedState` for multiple layer classification.
func NewPyState(modulePathName, moduleName, className string, params data.Map) (
	*PyState, error) {
	ins, err := newPyInstance(modulePathName, moduleName, className,
		[]data.Value{params}...)
	if err != nil {
		return nil, err
	}

	return &PyState{
		modulePath: modulePathName,
		moduleName: moduleName,
		className:  className,
		ins:        ins,
	}, nil
}

// newPyInstance creates a new Python class instance.
// User must call DecRef method to release a resource.
func newPyInstance(modulePathName, moduleName, className string, args ...data.Value) (
	py.ObjectInstance, error) {
	var null py.ObjectInstance
	py.ImportSysAndAppendPath(modulePathName)

	mdl, err := py.LoadModule(moduleName)
	if err != nil {
		return null, err
	}
	defer mdl.DecRef()

	ins, err := mdl.NewInstance(className, args...)
	if err != nil {
		return null, err
	}

	return ins, nil
}

// Terminate this state.
func (s *PyState) Terminate(ctx *core.Context) error {
	s.ins.DecRef()
	return nil
}

// Write and call "fit" function. Tuples is cached per train batch size.
func (s *PyState) Write(ctx *core.Context, t *core.Tuple) error {
	// TODO implement
	return nil
}

// PyMLPredict predicts data and return estimate value.
func PyMLPredict(ctx *core.Context, stateName string, dt data.Value) (data.Value, error) {
	s, err := lookupPyState(ctx, stateName)
	if err != nil {
		return nil, err
	}

	return s.ins.Call("predict", dt)
}

func lookupPyState(ctx *core.Context, stateName string) (*PyState, error) {
	st, err := ctx.SharedStates.Get(stateName)
	if err != nil {
		return nil, err
	}

	if s, ok := st.(*PyState); ok {
		return s, nil
	}

	return nil, fmt.Errorf("state '%v' isn't a PyState", stateName)
}

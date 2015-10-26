package pystate

import (
	"fmt"
	py "pfi/sensorbee/py/p"
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
)

// PyState is python constructor.
type PyState struct {
	modulePath    string
	moduleName    string
	className     string
	writeFuncName string

	ins py.ObjectInstance
}

// New creates `core.SharedState` for python constructor.
func New(modulePathName, moduleName, className string, writeFuncName string,
	params data.Map) (*PyState, error) {
	var ins py.ObjectInstance
	var err error
	if len(params) == 0 {
		ins, err = newPyInstance(modulePathName, moduleName, className)
	} else {
		ins, err = newPyInstance(modulePathName, moduleName, className,
			[]data.Value{params}...)
	}
	if err != nil {
		return nil, err
	}

	return &PyState{
		modulePath:    modulePathName,
		moduleName:    moduleName,
		className:     className,
		writeFuncName: writeFuncName,
		ins:           ins,
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

// Write calls "write" function.
// TODO should discuss this feature, bucket will be support?
func (s *PyState) Write(ctx *core.Context, t *core.Tuple) error {
	if s.writeFuncName == "" {
		return fmt.Errorf("state is not applied for writable")
	}
	_, err := s.ins.Call(s.writeFuncName, t.Data)
	return err
}

// Func calls instance method and return value.
func Func(ctx *core.Context, stateName string, funcName string, dt ...data.Value) (
	data.Value, error) {
	s, err := lookupPyState(ctx, stateName)
	if err != nil {
		return nil, err
	}

	return s.ins.Call(funcName, dt...)
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

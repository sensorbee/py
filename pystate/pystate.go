package pystate

import (
	"fmt"
	"pfi/sensorbee/py"
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
	"sync"
)

type PyState interface {
	core.SharedState
	lock()
	unlock()
	call(name string, args ...data.Value) (data.Value, error)
}

type pyState struct {
	modulePath string
	moduleName string
	className  string

	ins py.ObjectInstance

	mu sync.RWMutex
}

type pyWritableState struct {
	pyState
	writeFuncName string
}

func (s *pyState) lock() {
	s.mu.Lock()
}

func (s *pyState) unlock() {
	s.mu.Unlock()
}

func (s *pyState) call(name string, args ...data.Value) (data.Value, error) {
	return s.ins.Call(name, args...)
}

// New creates `core.SharedState` for python constructor.
func New(modulePathName, moduleName, className string, writeFuncName string,
	params data.Map) (PyState, error) {

	ins, err := newPyInstance(modulePathName, moduleName, className, params)
	if err != nil {
		return nil, err
	}

	state := pyState{
		modulePath: modulePathName,
		moduleName: moduleName,
		className:  className,
		ins:        ins,
	}
	// check if we have a writable state
	if writeFuncName != "" {
		return &pyWritableState{
			state,
			writeFuncName,
		}, nil
	}
	return &state, nil
}

// newPyInstance creates a new Python class instance.
// User must call DecRef method to release a resource.
func newPyInstance(modulePathName, moduleName, className string, args data.Map) (
	py.ObjectInstance, error) {
	var null py.ObjectInstance
	py.ImportSysAndAppendPath(modulePathName)

	mdl, err := py.LoadModule(moduleName)
	if err != nil {
		return null, err
	}
	defer mdl.DecRef()

	ins, err := mdl.NewInstanceWithKwd(className, args)
	if err != nil {
		return null, err
	}

	return ins, nil
}

// Terminate this state.
func (s *pyState) Terminate(ctx *core.Context) error {
	s.ins.DecRef()
	return nil
}

// Write calls "write" function.
// TODO should discuss this feature, bucket will be support?
func (s *pyWritableState) Write(ctx *core.Context, t *core.Tuple) error {
	if s.writeFuncName == "" {
		return fmt.Errorf("state is not applied for writable")
	}
	s.lock()
	defer s.unlock()

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
	s.lock()
	defer s.unlock()

	return s.call(funcName, dt...)
}

func lookupPyState(ctx *core.Context, stateName string) (PyState, error) {
	st, err := ctx.SharedStates.Get(stateName)
	if err != nil {
		return nil, err
	}

	if s, ok := st.(PyState); ok {
		return s, nil
	}

	return nil, fmt.Errorf("state '%v' isn't a PyState", stateName)
}

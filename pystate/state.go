package pystate

import (
	"gopkg.in/sensorbee/sensorbee.v0/core"
	"gopkg.in/sensorbee/sensorbee.v0/data"
	"io"
	"sync"
)

// state is a wrapper of a UDS written in Python. State is save/loadable,
// but doesn't support Write.
type state struct {
	// base is the base implementation of the state. Don't Embed this so that
	// users of State cannot directly call methods of BaseState.
	base *Base
	rwm  sync.RWMutex
}

// New creates `core.SharedState` for python constructor.
func New(baseParams *BaseParams, params data.Map) (core.SharedState, error) {
	bs, err := NewBase(baseParams, params)
	if err != nil {
		return nil, err
	}

	state := state{
		base: bs,
	}
	// check if we have a writable state
	if bs.params.WriteMethodName != "" {
		return &writableState{
			// Although this copies a RWMutex, the mutex isn't being locked at
			// the moment and it's safe to copy it now.
			state: state,
		}, nil
	}
	return &state, nil
}

func (s *state) Terminate(ctx *core.Context) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.base.Terminate(ctx)
}

func (s *state) Call(funcName string, dt ...data.Value) (data.Value, error) {
	// See BaseState.Call's godoc comment for the reason of using RLock here.
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return s.base.Call(funcName, dt...)
}

func (s *state) Save(ctx *core.Context, w io.Writer, params data.Map) error {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return s.base.Save(ctx, w, params)
}

func (s *state) Load(ctx *core.Context, r io.Reader, params data.Map) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	return s.base.Load(ctx, r, params)
}

// writableState is essentially same as state except its Write method support.
type writableState struct {
	state
}

func (s *writableState) Write(ctx *core.Context, t *core.Tuple) error {
	// See BaseState.Call's godoc comment for the reason of using RLock here.
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return s.base.Write(ctx, t)
}

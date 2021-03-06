package pystate

import (
	"gopkg.in/sensorbee/sensorbee.v0/bql/udf"
	"gopkg.in/sensorbee/sensorbee.v0/core"
	"gopkg.in/sensorbee/sensorbee.v0/data"
	"io"
)

// Creator is used by BQL to create state as a UDS.
type Creator struct {
}

var _ udf.UDSLoader = &Creator{}

// CreateState creates `core.SharedState` of a UDS written in Python.
// If params has parameters other than the ones defined in BaseParams, they
// will be directly passed to 'create' static method of the Python UDS.
func (c *Creator) CreateState(ctx *core.Context, params data.Map) (
	core.SharedState, error) {
	bp, err := ExtractBaseParams(params, true)
	if err != nil {
		return nil, err
	}
	return New(bp, params)
}

// LoadState loads saved state and creates a new instance from it.
func (c *Creator) LoadState(ctx *core.Context, r io.Reader, params data.Map) (
	core.SharedState, error) {
	base, err := LoadBase(ctx, r, params)
	if err != nil {
		return nil, err
	}
	s := state{
		base: base,
	}

	if s.base.params.WriteMethodName != "" {
		return &writableState{
			// Although this copies a RWMutex, the mutex isn't being locked at
			// the moment and it's safe to copy it now.
			state: s,
		}, nil
	}
	return &s, nil
}

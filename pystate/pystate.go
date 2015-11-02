package pystate

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ugorji/go/codec"
	"io"
	"io/ioutil"
	"os"
	"pfi/sensorbee/py"
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
	"sync"
)

// ErrAlreadyTerminated is occurred when called some python method after the
// SharedState is terminated.
var ErrAlreadyTerminated = errors.New("PyState is already terminated")

// PyState is a `SharedState` for python instance.
type PyState interface {
	core.SharedState
	Call(name string, args ...data.Value) (data.Value, error)
}

// State is a wrapper of a UDS written in Python. State is save/loadable,
// but doesn't support Write.
type State struct {
	modulePath      string
	moduleName      string
	className       string
	writeMethodName string

	ins *py.ObjectInstance

	rwm sync.RWMutex
}

// WritableState is essentially same as State except its Write method support.
type WritableState struct {
	State
}

type pyStateMsgpack struct {
	ModulePath      string `codec:"module_path"`
	ModuleName      string `codec:"module_name"`
	ClassName       string `codec:"class_name"`
	WriteMethodName string `codec:"write_method"`
}

// New creates `core.SharedState` for python constructor.
func New(modulePathName, moduleName, className string, writeMethodName string,
	params data.Map) (PyState, error) {

	var ins py.ObjectInstance
	var err error
	if params == nil || len(params) == 0 {
		ins, err = newPyInstance("create", modulePathName, moduleName, className)
	} else {
		ins, err = newPyInstance("create", modulePathName, moduleName, className,
			params)
	}
	if err != nil {
		return nil, err
	}

	state := State{}
	state.set(ins, modulePathName, moduleName, className, writeMethodName)
	// check if we have a writable state
	if writeMethodName != "" {
		return &WritableState{
			// Although this copies a RWMutex, the mutex isn't being locked at
			// the moment and it's safe to copy it now.
			state,
		}, nil
	}
	return &state, nil
}

// newPyInstance creates a new Python class instance.
// User must call DecRef method to release a resource.
func newPyInstance(createMethodName, modulePathName, moduleName, className string,
	args ...data.Value) (py.ObjectInstance, error) {
	var null py.ObjectInstance
	py.ImportSysAndAppendPath(modulePathName)

	mdl, err := py.LoadModule(moduleName)
	if err != nil {
		return null, err
	}
	defer mdl.DecRef()

	class, err := mdl.GetClass(className)
	if err != nil {
		return null, err
	}
	defer class.DecRef()

	var ins py.Object
	if args == nil || len(args) == 0 {
		ins, err = class.CallDirect(createMethodName)
	} else {
		ins, err = class.CallDirect(createMethodName, args...)
	}
	return py.ObjectInstance{ins}, err
}

func (s *State) set(ins py.ObjectInstance, modulePathName, moduleName,
	className, writeMethodName string) {
	if s.ins != nil {
		s.ins.DecRef()
	}

	s.modulePath = modulePathName
	s.moduleName = moduleName
	s.className = className
	s.writeMethodName = writeMethodName
	s.ins = &ins
}

// Terminate terminates the state.
func (s *State) Terminate(ctx *core.Context) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	if s.ins == nil {
		return nil // This isn't an error in Terminate
	}
	s.ins.DecRef()
	s.ins = nil
	return nil
}

// Write calls "write" function of the Python UDS.
func (s *WritableState) Write(ctx *core.Context, t *core.Tuple) error {
	// Although this write may modify the state of the Python UDS, it doesn't
	// change this State Go instance itself. Therefore, RLock is fine here.
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	if s.ins == nil {
		return ErrAlreadyTerminated
	}

	_, err := s.ins.Call(s.writeMethodName, t.Data)
	return err
}

// Call calls an instance method and returns its value.
func (s *State) Call(funcName string, dt ...data.Value) (data.Value, error) {
	// Although this call may modify the state of the Python UDS, it doesn't
	// change this State Go instance itself. Therefore, RLock is fine here.
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	if s.ins == nil {
		return nil, ErrAlreadyTerminated
	}

	return s.ins.Call(funcName, dt...)
}

// CallMethod calls an instance method and returns its value.
func CallMethod(ctx *core.Context, stateName, funcName string, dt ...data.Value) (
	data.Value, error) {
	s, err := lookupPyState(ctx, stateName)
	if err != nil {
		return nil, err
	}

	return s.Call(funcName, dt...)
}

// Save saves the model of the state. It saves its internal state and also calls
// 'save' method of the Python UDS. The Python UDS must save all the information
// necessary to reconstruct the current state including parameters passed by
// CREATE STATE statement.
func (s *State) Save(ctx *core.Context, w io.Writer, params data.Map) error {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	if s.ins == nil {
		return ErrAlreadyTerminated
	}

	if err := s.savePyMsgpack(w); err != nil {
		return err
	}

	temp, err := ioutil.TempFile("", "sensorbee_py_ml_state") // TODO: TempDir should be configurable
	if err != nil {
		return fmt.Errorf("cannot create a temporary file for saving data: %v",
			err)
	}
	filepath := temp.Name()
	if err := temp.Close(); err != nil {
		ctx.ErrLog(err).WithField("filepath", filepath).Warn(
			"Cannot close the temporary file")
	}
	defer func() {
		if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
			ctx.ErrLog(err).WithField("filepath", filepath).Warn(
				"Cannot remove the temporary file")
		}
	}()

	_, err = s.ins.Call("save", data.String(filepath), params)
	if err != nil {
		return err
	}

	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf(
			"cannot open the temporary file having the saved data: %v", err)
	}
	defer func() {
		if err := temp.Close(); err != nil {
			ctx.ErrLog(err).WithField("filepath", filepath).Warn(
				"Cannot close the temporary file")
		}
	}()
	_, err = io.Copy(w, f)
	return err
}

const (
	pyMLStateFormatVersion uint8 = 1
)

func (s *State) savePyMsgpack(w io.Writer) error {
	if _, err := w.Write([]byte{pyMLStateFormatVersion}); err != nil {
		return err
	}

	// Save parameter of PyMLState before save python's model
	save := &pyStateMsgpack{
		ModulePath:      s.modulePath,
		ModuleName:      s.moduleName,
		ClassName:       s.className,
		WriteMethodName: s.writeMethodName,
	}

	msgpackHandle := &codec.MsgpackHandle{}
	var out []byte
	enc := codec.NewEncoderBytes(&out, msgpackHandle)
	if err := enc.Encode(save); err != nil {
		return err
	}

	// Write size of pyMLMsgpack
	dataSize := uint32(len(out))
	err := binary.Write(w, binary.LittleEndian, dataSize)
	if err != nil {
		return err
	}

	// Write pyMLMsgpack in msgpack
	n, err := w.Write(out)
	if err != nil {
		return err
	}

	if n < len(out) {
		return errors.New("cannot save the pyMLMsgpack data")
	}

	return nil
}

// Load loads the model of the state. It reads the header of the saved file and
// calls 'load' static method of the Python UDS. 'load' static method creates
// a new instance of the Python UDS.
func (s *State) Load(ctx *core.Context, r io.Reader, params data.Map) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	if s.ins == nil {
		return ErrAlreadyTerminated
	}

	var formatVersion uint8
	if err := binary.Read(r, binary.LittleEndian, &formatVersion); err != nil {
		return err
	}

	// TODO: remove PyMLState specific parameters from params

	switch formatVersion {
	case 1:
		return s.loadPyMsgpackAndDataV1(ctx, r, params)
	default:
		return fmt.Errorf("unsupported format version of PyMLState container: %v",
			formatVersion)
	}
}

func (s *State) loadPyMsgpackAndDataV1(ctx *core.Context, r io.Reader,
	params data.Map) error {
	var dataSize uint32
	if err := binary.Read(r, binary.LittleEndian, &dataSize); err != nil {
		return err
	}
	if dataSize == 0 {
		return errors.New("size of pyMLMsgpack must be greater than 0")
	}

	// Read pyMLMsgpack from reader
	buf := make([]byte, dataSize)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	if n != int(dataSize) {
		return errors.New("read size is different from pyMLMsgpack")
	}

	// Desirialize pyMLMsgpack
	var saved pyStateMsgpack
	msgpackHandle := &codec.MsgpackHandle{}
	dec := codec.NewDecoderBytes(buf, msgpackHandle)
	if err := dec.Decode(&saved); err != nil {
		return err
	}

	temp, err := ioutil.TempFile("", "sensorbee_py_ml_state") // TODO: TempDir should be configurable
	if err != nil {
		return fmt.Errorf(
			"cannot create a temporary file to store the data to be loaded: %v",
			err)
	}
	filepath := temp.Name()
	tempClosed := false
	closeTemp := func() {
		if tempClosed {
			return
		}
		if err := temp.Close(); err != nil {
			ctx.ErrLog(err).WithField("filepath", filepath).Warn(
				"Cannot close the temporary file")
		}
		tempClosed = true
	}
	defer func() {
		closeTemp()
		if err := os.Remove(filepath); err != nil {
			ctx.ErrLog(err).WithField("filepath", filepath).Warn(
				"Cannot remove the temporary file")
		}
	}()
	if _, err := io.Copy(temp, r); err != nil {
		return err
	}
	closeTemp()

	ins, err := newPyInstance("load", saved.ModulePath, saved.ModuleName,
		saved.ClassName, []data.Value{data.String(filepath), params}...)
	if err != nil {
		return err
	}

	// TODO: Support alternative load strategy.
	// Currently, this method first loads a new model, and then release the old one.
	// However, releasing the old model before loading the new model is sometimes
	// required to reduce memory consumption. It should be configurable.

	// Exchange instance in `s` when Load succeeded
	s.set(ins, saved.ModulePath, saved.ModuleName, saved.ClassName,
		saved.WriteMethodName)
	return nil
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

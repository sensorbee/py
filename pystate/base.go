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
)

// ErrAlreadyTerminated is occurred when called some python method after the
// SharedState is terminated.
var ErrAlreadyTerminated = errors.New("pystate is already terminated")

// BaseParams has parameters for Base given in WITH clause of CREATE STATE
// statement.
type BaseParams struct {
	// ModulePath is a path at where the target Python module is located.
	// This parameter can be set as "module_path" in a WITH clause. This is
	// a required parameter.
	ModulePath string `codec:"module_path"`

	// ModuleName is a name of a Python module to be loaded. This parameter
	// can be set as "module_name" in a WITH clause. This is a required parameter.
	ModuleName string `codec:"module_name"`

	// ClassName is a name of a class in the Python module to be loaded. This
	// parameter can be set as "class_name" in a WITH clause. This is a required
	// parameter.
	ClassName string `codec:"class_name"`

	// WriteMethodName is a name of a method which handles Write calls, which
	// is mostly done by 'uds' Sink. When this parameter is specified, a UDS
	// will be writable. Otherwise, it doesn't support Write.
	WriteMethodName string `codec:"write_method"`
}

var (
	modulePath      = data.MustCompilePath("module_path")
	moduleNamePath  = data.MustCompilePath("module_name")
	classNamePath   = data.MustCompilePath("class_name")
	writeMethodPath = data.MustCompilePath("write_method")
)

// ExtractBaseParams extract parameters for Base from parameters given in
// a WITH clause of a CREATE STATE statement. If removeBaseKeys is true,
// this function removes base parameters from params and only other parameters
// remain in the map when this function succeeds. If this function fails,
// all parameters including base parameters remain in the map.
func ExtractBaseParams(params data.Map, removeBaseKeys bool) (*BaseParams, error) {
	bp := &BaseParams{}

	if mp, err := params.Get(modulePath); err == nil {
		if p, err := data.AsString(mp); err != nil {
			return nil, err
		} else {
			bp.ModulePath = p
		}
	}

	if mn, err := params.Get(moduleNamePath); err != nil {
		return nil, err
	} else if moduleName, err := data.AsString(mn); err != nil {
		return nil, err
	} else {
		bp.ModuleName = moduleName
	}

	if cn, err := params.Get(classNamePath); err != nil {
		return nil, err
	} else if className, err := data.AsString(cn); err != nil {
		return nil, err
	} else {
		bp.ClassName = className
	}

	if wmn, err := params.Get(writeMethodPath); err == nil {
		bp.WriteMethodName, err = data.AsString(wmn)
		if err != nil {
			return nil, err
		}
	}

	if removeBaseKeys {
		for _, k := range []string{"module_path", "module_name", "class_name", "write_method"} {
			delete(params, k)
		}
	}
	return bp, nil
}

// Base is a wrapper of a UDS written in Python. It has common implementations
// that can be shared with State, WritableState, and other wrappers. Base
// doesn't acquire lock and the caller should provide concurrency control
// over them. Each method describes what kind of lock it requires.
type Base struct {
	params BaseParams
	ins    *py.ObjectInstance
}

// NewBase creates a new Base state.
func NewBase(baseParams *BaseParams, params data.Map) (*Base, error) {
	var (
		ins py.ObjectInstance
		err error
	)

	ins, err = newPyInstance("create", baseParams, nil, params)
	if err != nil {
		return nil, err
	}

	s := Base{}
	s.set(ins, baseParams)
	return &s, nil
}

// newPyInstance creates a new Python class instance.
// User must call DecRef method to release a resource.
func newPyInstance(createMethodName string, baseParams *BaseParams,
	args []data.Value, kwdArgs data.Map) (py.ObjectInstance, error) {
	var null py.ObjectInstance
	py.ImportSysAndAppendPath(baseParams.ModulePath)

	mdl, err := py.LoadModule(baseParams.ModuleName)
	if err != nil {
		return null, err
	}
	defer mdl.DecRef()

	class, err := mdl.GetClass(baseParams.ClassName)
	if err != nil {
		return null, err
	}
	defer class.DecRef()

	ins, err := class.CallDirect(createMethodName, nil, kwdArgs)
	return py.ObjectInstance{ins}, err
}

func (s *Base) set(ins py.ObjectInstance, baseParams *BaseParams) {
	if s.ins != nil {
		s.ins.DecRef()
	}
	s.params = *baseParams
	s.ins = &ins
}

// Terminate terminates the state.
//
// This method requires write-lock.
func (s *Base) Terminate(ctx *core.Context) error {
	if s.ins == nil {
		return nil // This isn't an error in Terminate
	}
	s.ins.DecRef()
	s.ins = nil
	return nil
}

// CheckTermination checks if the Base is already terminated. It returns nil
// if the Base is still working. It returns ErrAlreadyTerminated if the Base
// has already been terminated.
//
// This method requires read-lock.
func (s *Base) CheckTermination() error {
	if s.ins == nil {
		return ErrAlreadyTerminated
	}
	return nil
}

// Call calls an instance method and returns its value.
//
// Although this call may modify the state of the Python UDS, it doesn't
// change this Base Go instance itself. Therefore, this method requires
// read-lock.
func (s *Base) Call(funcName string, dt ...data.Value) (data.Value, error) {
	if s.ins == nil {
		return nil, ErrAlreadyTerminated
	}
	return s.ins.Call(funcName, dt...)
}

// Write calls "write" function of the Python UDS.
//
// Although this write may modify the state of the Python UDS, it doesn't
// change this Base Go instance itself. Therefore, RLock is fine here.
func (s *Base) Write(ctx *core.Context, t *core.Tuple) error {
	if s.ins == nil {
		return ErrAlreadyTerminated
	}
	_, err := s.ins.Call(s.params.WriteMethodName, t.Data)
	return err
}

// Save saves the model of the state. It saves its internal state and also calls
// 'save' method of the Python UDS. The Python UDS must save all the information
// necessary to reconstruct the current state including parameters passed by
// CREATE STATE statement.
//
// This method requires read-Lock.
func (s *Base) Save(ctx *core.Context, w io.Writer, params data.Map) error {
	if s.ins == nil {
		return ErrAlreadyTerminated
	}

	if err := s.savePyMsgpack(w); err != nil {
		return err
	}

	temp, err := ioutil.TempFile("", "sensorbee_py_state") // TODO: TempDir should be configurable
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
	pyBaseStateFormatVersion uint8 = 1
)

func (s *Base) savePyMsgpack(w io.Writer) error {
	if _, err := w.Write([]byte{pyBaseStateFormatVersion}); err != nil {
		return err
	}

	// Save parameter of Base before save python's model
	msgpackHandle := &codec.MsgpackHandle{}
	var out []byte
	enc := codec.NewEncoderBytes(&out, msgpackHandle)
	if err := enc.Encode(&s.params); err != nil {
		return err
	}

	// Write size of BaseParams
	dataSize := uint32(len(out))
	err := binary.Write(w, binary.LittleEndian, dataSize)
	if err != nil {
		return err
	}

	// Write BaseParams in msgpack
	n, err := w.Write(out)
	if err != nil {
		return err
	}

	if n < len(out) {
		return errors.New("cannot save the BaseParams data")
	}

	return nil
}

// Load loads the model of the state. It reads the header of the saved file and
// calls 'load' static method of the Python UDS. 'load' static method creates
// a new instance of the Python UDS.
//
// This method requires write-lock.
func (s *Base) Load(ctx *core.Context, r io.Reader, params data.Map) error {
	if s.ins == nil {
		return ErrAlreadyTerminated
	}

	var formatVersion uint8
	if err := binary.Read(r, binary.LittleEndian, &formatVersion); err != nil {
		return err
	}

	// TODO: remove PyStateState specific parameters from params

	switch formatVersion {
	case 1:
		return s.loadPyMsgpackAndDataV1(ctx, r, params)
	default:
		return fmt.Errorf("unsupported format version of pystate container: %v",
			formatVersion)
	}
}

func (s *Base) loadPyMsgpackAndDataV1(ctx *core.Context, r io.Reader,
	params data.Map) error {
	var dataSize uint32
	if err := binary.Read(r, binary.LittleEndian, &dataSize); err != nil {
		return err
	}
	if dataSize == 0 {
		return errors.New("size of BaseParams must be greater than 0")
	}

	// Read BaseParams from reader
	buf := make([]byte, dataSize)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	if n != int(dataSize) {
		return errors.New("read size is different from BaseParams")
	}

	// Desirialize BaseParams
	var saved BaseParams
	msgpackHandle := &codec.MsgpackHandle{}
	dec := codec.NewDecoderBytes(buf, msgpackHandle)
	if err := dec.Decode(&saved); err != nil {
		return err
	}

	temp, err := ioutil.TempFile("", "sensorbee_py_state") // TODO: TempDir should be configurable
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

	ins, err := newPyInstance("load", &saved, []data.Value{data.String(filepath)},
		params)
	if err != nil {
		return err
	}

	// TODO: Support alternative load strategy.
	// Currently, this method first loads a new model, and then release the old one.
	// However, releasing the old model before loading the new model is sometimes
	// required to reduce memory consumption. It should be configurable.

	// Exchange instance in `s` when Load succeeded
	s.set(ins, &saved)
	return nil
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

type pyState interface {
	core.SharedState
	Call(funcName string, dt ...data.Value) (data.Value, error)
}

func lookupPyState(ctx *core.Context, stateName string) (pyState, error) {
	st, err := ctx.SharedStates.Get(stateName)
	if err != nil {
		return nil, err
	}

	if s, ok := st.(pyState); ok {
		return s, nil
	}

	return nil, fmt.Errorf("state '%v' isn't a pystate.State or pystate.WritableState", stateName)
}

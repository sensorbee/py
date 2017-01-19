// +build py3.4 py3.5 py3.6

package py

func loadExceptionModule() (ObjectModule, error) {
	return LoadModule("builtins")
}

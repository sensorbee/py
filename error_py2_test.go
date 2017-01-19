// +build !py3.4
// +build !py3.5
// +build !py3.6

package py

func loadExceptionModule() (ObjectModule, error) {
	return LoadModule("exceptions")
}

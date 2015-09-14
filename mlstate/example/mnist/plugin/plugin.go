package plugin

import (
	"pfi/sensorbee/pystate/mlstate/example/mnist"
	"pfi/sensorbee/sensorbee/bql"
)

func init() {
	bql.MustRegisterGlobalSourceCreator("mnist_source", &mnist.DataSourceCreator{})
}

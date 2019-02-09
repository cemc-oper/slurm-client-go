package common

import "github.com/perillaroc/nwpc-hpc-model-go"

type ItemFilter interface {
	Apply(items []hpcmodel.Item) []hpcmodel.Item
	IsUseInDefault() bool
}

package common

import "github.com/cemc-oper/hpc-model-go"

type ItemFilter interface {
	Apply(items []hpcmodel.Item) []hpcmodel.Item
	IsUseInDefault() bool
}

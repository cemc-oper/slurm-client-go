package long_time_job

import (
	"github.com/cemc-oper/hpc-model-go"
	"time"
)

type LongTimeJobFilter struct {
	users    []string
	duration time.Duration
}

func (f *LongTimeJobFilter) IsUseInDefault() bool {
	return true
}

func (f *LongTimeJobFilter) Apply(items []hpcmodel.Item) []hpcmodel.Item {
	dateCondition := hpcmodel.DateTimePropertyFilterCondition{
		ID: "squeue.submit_time",
		Checker: &hpcmodel.DateTimeBeforeValueChecker{
			ExpectedValue: time.Now().Add(-1 * f.duration),
		},
	}

	ownerCondition := hpcmodel.StringPropertyFilterCondition{
		ID: "squeue.owner",
		Checker: &hpcmodel.StringInValueChecker{
			ExpectedValues: f.users,
		},
	}

	filter := hpcmodel.Filter{
		Conditions: []hpcmodel.FilterCondition{
			&ownerCondition,
			&dateCondition,
		},
	}
	targetItems := filter.Filter(items)
	return targetItems
}

func CreateFilter() *LongTimeJobFilter {
	return &LongTimeJobFilter{
		users:    []string{"op_gfs", "op_meso", "op_eps", "op_spec", "op_post", "op_srv"},
		duration: time.Hour * 5,
	}
}

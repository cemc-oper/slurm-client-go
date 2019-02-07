package common_test

import (
	"github.com/perillaroc/nwpc-hpc-model-go"
	"slurm-client-go/common"
	"strings"
	"testing"
)

func TestGetSqueueQueryModel(t *testing.T) {
	line := `ACCOUNT|TRES_PER_NODEMIN_CPUS|MIN_TMP_DISK|END_TIME|FEATURES|GROUP|OVER_SUBSCRIBE|JOBID|NAME|COMMENT|TIME_LIMIT|MIN_MEMORY|REQ_NODES|COMMAND|PRIORITY|QOS|REASON||ST|USER|RESERVATION|WCKEY|EXC_NODES|NICE|S:C:T|JOBID|EXEC_HOST|CPUS|NODES|DEPENDENCY|ARRAY_JOB_ID|GROUP|SOCKETS_PER_NODE|CORES_PER_SOCKET|THREADS_PER_CORE|ARRAY_TASK_ID|TIME_LEFT|TIME|NODELIST|CONTIGUOUS|PARTITION|PRIORITY|NODELIST(REASON)|START_TIME|STATE|UID|SUBMIT_TIME|LICENSES|CORE_SPEC|SCHEDNODES|WORK_DIR
chendh|N/A32|0|2019-02-13T20:30:06|(null)|nwpc|OK|5836542|GRAPES|GRAPES|15-00:00:00|5G||/g8/JOB_TMP/chendh/ShCu/RUN_24/grapes.sbatch|0.00000011641532|middle|None||R|chendh|(null)|(null)||0|*:*:*|5836542|cmac0819|1024|32||5836542|10201|*|*|*|N/A|12-11:23:44|2-12:36:16|cmac[0819-0832,0835-0850,0897-0898]|0|normal|500|cmac[0819-0832,0835-0850,0897-0898]|2019-01-29T20:30:06|RUNNING|1020104|2019-01-29T08:41:45|(null)|N/A|(null)|/g8/JOB_TMP/chendh/ShCu/RUN_24
lijl|N/A32|0|2019-02-13T00:37:51|(null)|csmd|OK|5831234|BCC800|BCCCSM|15-00:00:00|5G||/g6/lijl/BCC_CSMv3.v20190124/p25_2/build.csh|0.00000011641532|normal|None||R|lijl|(null)|(null)||0|*:*:*|5831234|cmac0243|704|22||5831234|10304|*|*|*|N/A|11-15:31:29|3-08:28:31|cmac[0243-0248,1155-1170]|0|normal|500|cmac[0243-0248,1155-1170]|2019-01-29T00:37:51|RUNNING|1030410|2019-01-29T00:37:49|(null)|N/A|(null)|/g6/lijl/BCC_CSMv3.v20190124/p25_2
wutw|N/A32|0|2019-02-07T10:28:35|(null)|csmd|OK|5720780|BCCCSM|WRF|15-00:00:00|5G||/g6/wutw/BCC_CSMv3_T266p13_OK/T266L56.deep.955/build.csh|0.00000011641532|normal|None||R|wutw|(null)|(null)||0|*:*:*|5720780|cmac0272|576|18||5720780|10304|*|*|*|N/A|6-01:22:13|8-22:37:47|cmac[0272-0273,0499-0514]|0|normal|500|cmac[0272-0273,0499-0514]|2019-01-23T10:28:35|RUNNING|1030402|2019-01-23T10:28:35|(null)|N/A|(null)|/g6/wutw/BCC_CSMv3_T266p13_OK/T266L56.deep.955`
	lines := strings.Split(line, "\n")
	model, err := common.GetSqueueQueryModel(lines)
	if err != nil {
		t.Errorf("get query model failed: %v", err)
	}

	tests := []struct {
		index int
		id    string
		data  string
	}{
		{
			0,
			"squeue.account",
			"chendh",
		},
		{
			0,
			"squeue.partition",
			"normal",
		},
		{
			2,
			"squeue.account",
			"wutw",
		},
	}

	for _, test := range tests {
		item := model.Items[test.index]
		prop := item.GetProperty(test.id).(*hpcmodel.StringProperty)
		if prop == nil {
			t.Errorf("Item %d property %s not found",
				test.index, test.id)
		}
		if prop.Data != test.data {
			t.Errorf("Item %d property %s data is %s, required %s",
				test.index, test.id, prop.Data, test.data)
		}
	}
}

func TestGetSqueueQueryModelFilter(t *testing.T) {
	line := `ACCOUNT|TRES_PER_NODEMIN_CPUS|MIN_TMP_DISK|END_TIME|FEATURES|GROUP|OVER_SUBSCRIBE|JOBID|NAME|COMMENT|TIME_LIMIT|MIN_MEMORY|REQ_NODES|COMMAND|PRIORITY|QOS|REASON||ST|USER|RESERVATION|WCKEY|EXC_NODES|NICE|S:C:T|JOBID|EXEC_HOST|CPUS|NODES|DEPENDENCY|ARRAY_JOB_ID|GROUP|SOCKETS_PER_NODE|CORES_PER_SOCKET|THREADS_PER_CORE|ARRAY_TASK_ID|TIME_LEFT|TIME|NODELIST|CONTIGUOUS|PARTITION|PRIORITY|NODELIST(REASON)|START_TIME|STATE|UID|SUBMIT_TIME|LICENSES|CORE_SPEC|SCHEDNODES|WORK_DIR
chendh|N/A32|0|2019-02-13T20:30:06|(null)|nwpc|OK|5836542|GRAPES|GRAPES|15-00:00:00|5G||/g8/JOB_TMP/chendh/ShCu/RUN_24/grapes.sbatch|0.00000011641532|middle|None||R|chendh|(null)|(null)||0|*:*:*|5836542|cmac0819|1024|32||5836542|10201|*|*|*|N/A|12-11:23:44|2-12:36:16|cmac[0819-0832,0835-0850,0897-0898]|0|normal|500|cmac[0819-0832,0835-0850,0897-0898]|2019-01-29T20:30:06|RUNNING|1020104|2019-01-29T08:41:45|(null)|N/A|(null)|/g8/JOB_TMP/chendh/ShCu/RUN_24
lijl|N/A32|0|2019-02-13T00:37:51|(null)|csmd|OK|5831234|BCC800|BCCCSM|15-00:00:00|5G||/g6/lijl/BCC_CSMv3.v20190124/p25_2/build.csh|0.00000011641532|normal|None||R|lijl|(null)|(null)||0|*:*:*|5831234|cmac0243|704|22||5831234|10304|*|*|*|N/A|11-15:31:29|3-08:28:31|cmac[0243-0248,1155-1170]|0|normal|500|cmac[0243-0248,1155-1170]|2019-01-29T00:37:51|RUNNING|1030410|2019-01-29T00:37:49|(null)|N/A|(null)|/g6/lijl/BCC_CSMv3.v20190124/p25_2
wutw|N/A32|0|2019-02-07T10:28:35|(null)|csmd|OK|5720780|BCCCSM|WRF|15-00:00:00|5G||/g6/wutw/BCC_CSMv3_T266p13_OK/T266L56.deep.955/build.csh|0.00000011641532|normal|None||R|wutw|(null)|(null)||0|*:*:*|5720780|cmac0272|576|18||5720780|10304|*|*|*|N/A|6-01:22:13|8-22:37:47|cmac[0272-0273,0499-0514]|0|normal|500|cmac[0272-0273,0499-0514]|2019-01-23T10:28:35|RUNNING|1030402|2019-01-23T10:28:35|(null)|N/A|(null)|/g6/wutw/BCC_CSMv3_T266p13_OK/T266L56.deep.955`
	lines := strings.Split(line, "\n")
	model, err := common.GetSqueueQueryModel(lines)
	if err != nil {
		t.Errorf("get query model failed: %v", err)
	}

	filter := hpcmodel.Filter{}

	checker := hpcmodel.StringInValueChecker{
		ExpectedValues: []string{"wutw"},
	}
	condition := hpcmodel.StringPropertyFilterCondition{
		ID:      "squeue.account",
		Checker: &checker,
	}
	filter.Conditions = append(filter.Conditions, &condition)

	targetItems := filter.Filter(model.Items)
	if len(targetItems) != 1 {
		t.Errorf("filter error")
	}
}

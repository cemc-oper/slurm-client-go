package cmd

import (
	"fmt"
	hpcmodel "github.com/nwpc-oper/hpc-model-go"
	"github.com/nwpc-oper/slurm-client-go/common"
	"github.com/spf13/cobra"
	"log"
	"strconv"
	"strings"
	"time"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch jobs",
	Long:  "Watch jobs until finished.",
	Run: func(cmd *cobra.Command, args []string) {
		WatchCommand(watchUsers, watchPartitions, watchJobs)
	},
}

var watchUsers []string
var watchPartitions []string
var watchJobs []string

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.PersistentFlags().StringArrayVarP(
		&watchUsers, "user", "u", []string{}, "user")
	watchCmd.PersistentFlags().StringArrayVarP(
		&watchPartitions, "partition", "p", []string{}, "partition")
	watchCmd.PersistentFlags().StringArrayVarP(
		&watchJobs, "job", "j", []string{}, "jobs")
}

func getRunningJobs(users []string, partitions []string, jobs []string) []hpcmodel.Item {
	params := []string{"-o", "%all"}

	filter := hpcmodel.Filter{}

	if len(users) > 0 {
		checker := hpcmodel.StringInValueChecker{
			ExpectedValues: users,
		}
		condition := hpcmodel.StringPropertyFilterCondition{
			ID:      "squeue.user",
			Checker: &checker,
		}
		filter.Conditions = append(filter.Conditions, &condition)
	}

	if len(partitions) > 0 {
		checker := hpcmodel.StringInValueChecker{
			ExpectedValues: partitions,
		}
		condition := hpcmodel.StringPropertyFilterCondition{
			ID:      "squeue.partition",
			Checker: &checker,
		}
		filter.Conditions = append(filter.Conditions, &condition)
	}

	if len(jobs) > 0 {
		checker := hpcmodel.StringInValueChecker{
			ExpectedValues: jobs,
		}
		condition := hpcmodel.StringPropertyFilterCondition{
			ID:      "squeue.job_id",
			Checker: &checker,
		}
		filter.Conditions = append(filter.Conditions, &condition)
	}

	lines, err := common.GetSqueueCommandResult(params)
	if err != nil {
		log.Fatalf("get query result error: %v", err)
	}

	model, err := common.GetSqueueQueryModel(lines)
	if err != nil {
		log.Fatalf("model build failed: %v", err)
	}

	targetItems := filter.Filter(model.Items)
	return targetItems
}

func getCurrentTime() string {
	t := time.Now()
	return t.Format("2006-01-02 15:04:05")
}

func WatchCommand(users []string, partitions []string, jobs []string) {
	for true {
		items := getRunningJobs(users, partitions, jobs)
		jobLens := len(items)
		if jobLens == 0 {
			fmt.Printf("[%s]checking jobs...done\n", getCurrentTime())
			break
		}

		job_map := make(map[string]int)
		for _, item := range items {
			state := item.GetProperty("squeue.state").(*hpcmodel.StringProperty).Text
			val, ok := job_map[state]
			if ok {
				job_map[state] = val + 1
			} else {
				job_map[state] = 1
			}
		}
		var tokens []string
		for k, v := range job_map {
			tokens = append(tokens, k+": "+strconv.Itoa(v))
		}
		summary := strings.Join(tokens, ", ")

		fmt.Printf("[%s]checking jobs...%s\n", getCurrentTime(), summary)
		time.Sleep(time.Minute)
	}

}

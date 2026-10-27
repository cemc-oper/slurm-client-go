package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	hpcmodel "github.com/cemc-oper/hpc-model-go"
	"github.com/spf13/cobra"
)

type watchClock interface {
	Now() time.Time
	Sleep(duration time.Duration)
}

type realWatchClock struct{}

func (realWatchClock) Now() time.Time {
	return time.Now()
}

func (realWatchClock) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch jobs",
	Long:  "Watch jobs until finished.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return watchCommand(cmd.OutOrStdout(), slurmCommandRunner{}, realWatchClock{}, watchUsers, watchPartitions, watchJobs)
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

func WatchCommand(users []string, partitions []string, jobs []string) error {
	return watchCommand(os.Stdout, slurmCommandRunner{}, realWatchClock{}, users, partitions, jobs)
}

func watchCommand(output io.Writer, runner commandRunner, clock watchClock, users []string, partitions []string, jobs []string) error {
	for {
		items, err := fetchSqueueItems(runner, squeueQueryOptions{users: users, partitions: partitions, jobs: jobs})
		if err != nil {
			return err
		}
		currentTime := clock.Now().Format("2006-01-02 15:04:05")
		if len(items) == 0 {
			if _, err := fmt.Fprintf(output, "[%s]checking jobs...done\n", currentTime); err != nil {
				return err
			}
			return nil
		}

		stateCounts := make(map[string]int)
		for _, item := range items {
			state := item.GetProperty("squeue.state").(*hpcmodel.StringProperty).Text
			stateCounts[state]++
		}
		states := make([]string, 0, len(stateCounts))
		for state := range stateCounts {
			states = append(states, state)
		}
		sort.Strings(states)
		tokens := make([]string, 0, len(states))
		for _, state := range states {
			tokens = append(tokens, fmt.Sprintf("%s: %d", state, stateCounts[state]))
		}

		if _, err := fmt.Fprintf(output, "[%s]checking jobs...%s\n", currentTime, strings.Join(tokens, ", ")); err != nil {
			return err
		}
		clock.Sleep(time.Minute)
	}
}

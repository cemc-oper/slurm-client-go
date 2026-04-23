package cmd

import (
	"github.com/cemc-oper/slurm-client-go/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive TUI for Slurm jobs",
	Long:  "Launch an interactive terminal UI to browse and inspect Slurm jobs.",
	Run: func(cmd *cobra.Command, args []string) {
		tui.TUICommand()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

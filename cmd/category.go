package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/nwpc-oper/slurm-client-go/common"
	"github.com/spf13/cobra"
)

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "category list defined by command",
	Long:  "category list defined by command",
	Run: func(cmd *cobra.Command, args []string) {
		CategoryCommand()
	},
}

var categoryDetail bool

func init() {
	rootCmd.AddCommand(categoryCmd)
	categoryCmd.PersistentFlags().BoolVarP(&categoryDetail, "detail", "d", false,
		"show detail information")
}

func CategoryCommand() {
	categoryList := common.BuildSqueueCategoryList()

	boldColor := color.New(color.Bold).SprintFunc()
	blueColor := color.New(color.FgBlue).SprintfFunc()

	for _, category := range categoryList.CategoryList {
		fmt.Printf("%s\n", boldColor(category.ID))
		if !categoryDetail {
			continue
		}
		fmt.Printf("  %s: %s\n", blueColor("display name"), category.DisplayName)
		fmt.Printf("  %s: %s\n", blueColor("label"), category.Label)
		fmt.Printf("  %s: %s\n", blueColor("record parser"), category.RecordParserClass)
		fmt.Printf("  %s:\n", blueColor("record parser arguments"))
		for _, arg := range category.RecordParserArguments {
			fmt.Printf("    %s\n", arg)
		}
		fmt.Printf("  %s: %s\n", blueColor("property class"), category.PropertyClass)
		fmt.Printf("  %s:\n", blueColor("property create arguments"))
		for _, arg := range category.PropertyCreateArguments {
			fmt.Printf("    %s\n", arg)
		}
		fmt.Println()
	}
}

package cmd

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/cemc-oper/slurm-client-go/common"
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

	idStyle := lipgloss.NewStyle().Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#64B5F6"))

	for _, category := range categoryList.CategoryList {
		fmt.Printf("%s\n", idStyle.Render(category.ID))
		if !categoryDetail {
			continue
		}
		fmt.Printf("  %s: %s\n", labelStyle.Render("display name"), category.DisplayName)
		fmt.Printf("  %s: %s\n", labelStyle.Render("label"), category.Label)
		fmt.Printf("  %s: %s\n", labelStyle.Render("record parser"), category.RecordParserClass)
		fmt.Printf("  %s:\n", labelStyle.Render("record parser arguments"))
		for _, arg := range category.RecordParserArguments {
			fmt.Printf("    %s\n", arg)
		}
		fmt.Printf("  %s: %s\n", labelStyle.Render("property class"), category.PropertyClass)
		fmt.Printf("  %s:\n", labelStyle.Render("property create arguments"))
		for _, arg := range category.PropertyCreateArguments {
			fmt.Printf("    %s\n", arg)
		}
		fmt.Println()
	}
}

package common

import "github.com/cemc-oper/hpc-model-go"

// FetchSqueueItems runs squeue -o %all, parses the output, and returns job items for the TUI.
func FetchSqueueItems() ([]hpcmodel.Item, error) {
	params := []string{"-o", "%all"}
	lines, err := GetSqueueCommandResult(params)
	if err != nil {
		return nil, err
	}
	model, err := GetSqueueQueryModel(lines)
	if err != nil {
		return nil, err
	}
	return model.Items, nil
}

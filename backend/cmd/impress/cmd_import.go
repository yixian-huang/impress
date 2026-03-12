package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func importCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Import site data from JSON",
		Long:  "Import site content from a previously exported JSON file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("impress import: not yet implemented")
			return nil
		},
	}
}

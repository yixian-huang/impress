package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func exportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export site data to JSON",
		Long:  "Export all site content (articles, pages, settings) to a JSON file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("impress export: not yet implemented")
			return nil
		},
	}
}

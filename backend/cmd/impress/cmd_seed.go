package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func seedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Seed the database with sample data",
		Long:  "Populate the database with example content for development and demos.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("impress seed: not yet implemented")
			return nil
		},
	}
}

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Impress CMS project",
		Long:  "Interactive project initialization: choose database, port, and generate configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("impress init: not yet implemented")
			return nil
		},
	}
}

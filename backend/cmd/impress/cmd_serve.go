package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the Impress CMS development server",
		Long:  "Start the backend server with sensible defaults. Reads .env or environment variables.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("impress serve: not yet implemented")
			return nil
		},
	}
}

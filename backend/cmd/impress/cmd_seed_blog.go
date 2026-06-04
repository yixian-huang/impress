package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"blotting-consultancy/internal/repository"
	"blotting-consultancy/internal/seed"
)

func seedBlogSamplesCmd() *cobra.Command {
	var dsn string

	cmd := &cobra.Command{
		Use:   "blog-samples",
		Short: "Seed sample blog categories, tags, and articles",
		Long:  "Inserts ~48 published sample articles (idempotent by slug sample-post-XX) for local UI testing.",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := openDatabase(dsn)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}

			articleRepo := repository.NewGormArticleRepository(database.DB)
			categoryRepo := repository.NewGormCategoryRepository(database.DB)
			tagRepo := repository.NewGormTagRepository(database.DB)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			if err := seed.SeedBlogSamples(ctx, articleRepo, categoryRepo, tagRepo); err != nil {
				return fmt.Errorf("blog sample seed failed: %w", err)
			}

			fmt.Println("Blog sample data seeded successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&dsn, "dsn", "", "Database DSN (default: DB_DSN env var or SQLite)")
	return cmd
}

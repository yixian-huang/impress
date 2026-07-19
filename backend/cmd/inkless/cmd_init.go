package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/yixian-huang/inkless/backend/pkg/config"

	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	var nonInteractive bool
	var port, dbType, outputDir, runtimeEnv string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Inkless CMS project",
		Long:  "Interactive project initialization: choose database, port, and generate .env configuration file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				listenPort int
				dsn        string
				envName    string
				err        error
			)
			reader := bufio.NewReader(os.Stdin)

			if nonInteractive {
				listenPort, err = config.ParsePortString(port)
				if err != nil {
					return err
				}
				envName = runtimeEnv
				if dbType == "postgres" {
					dsn = "postgres://inkless:inkless@localhost:5432/inkless?sslmode=disable"
				} else {
					dsn, err = config.BuildDSN(config.DatabaseInput{Type: "sqlite", SQLitePath: "./data/inkless.db"})
					if err != nil {
						return err
					}
				}
			} else {
				fmt.Print("Port [8088]: ")
				input, _ := reader.ReadString('\n')
				listenPort, err = config.ParsePortString(strings.TrimSpace(input))
				if err != nil {
					return err
				}

				fmt.Print("Runtime ENV (development/production) [development]: ")
				input, _ = reader.ReadString('\n')
				envName, err = normalizeCLIEnv(strings.TrimSpace(input))
				if err != nil {
					return err
				}

				fmt.Print("Database (sqlite/postgres) [sqlite]: ")
				input, _ = reader.ReadString('\n')
				dbChoice := strings.TrimSpace(input)
				if dbChoice == "" || dbChoice == "sqlite" {
					dsn, err = config.BuildDSN(config.DatabaseInput{Type: "sqlite", SQLitePath: "./data/inkless.db"})
					if err != nil {
						return err
					}
				} else {
					fmt.Print("PostgreSQL DSN [postgres://inkless:inkless@localhost:5432/inkless?sslmode=disable]: ")
					input, _ = reader.ReadString('\n')
					dsn = strings.TrimSpace(input)
					if dsn == "" {
						dsn = "postgres://inkless:inkless@localhost:5432/inkless?sslmode=disable"
					}
				}
			}

			outDir := outputDir
			if outDir == "" {
				outDir = "."
			}

			secret, refresh, err := config.GenerateJWTSecrets()
			if err != nil {
				return err
			}

			envPath, err := config.WriteEnvFile(outDir, config.EnvFileParams{
				Port:             listenPort,
				DBDSN:            dsn,
				JWTSecret:        secret,
				JWTRefreshSecret: refresh,
				Env:              envName,
				UploadDir:        "./uploads",
			})
			if err != nil {
				return err
			}

			fmt.Printf("Configuration written to %s\n", envPath)
			fmt.Println("Directories created: data/, uploads/")
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Review and update .env (especially JWT secrets)")
			fmt.Println("  2. Run: inkless migrate up")
			fmt.Println("  3. Open /setup in the browser or run: inkless seed")
			fmt.Println("  4. Run: inkless serve")
			return nil
		},
	}

	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Skip interactive prompts")
	cmd.Flags().StringVar(&port, "port", "8088", "Server port (non-interactive mode)")
	cmd.Flags().StringVar(&dbType, "db", "sqlite", "Database type: sqlite or postgres (non-interactive mode)")
	cmd.Flags().StringVar(&runtimeEnv, "env", "development", "Runtime ENV written to .env")
	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for config files")

	return cmd
}

func normalizeCLIEnv(raw string) (string, error) {
	if raw == "" {
		return "development", nil
	}
	switch strings.ToLower(raw) {
	case "development", "production":
		return strings.ToLower(raw), nil
	default:
		return "", fmt.Errorf("env must be development or production")
	}
}

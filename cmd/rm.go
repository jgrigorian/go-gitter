package cmd

import (
	"fmt"

	"github.com/jgrigorian/go-gitter/internal/config"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <path-or-name>",
	Short: "Remove a repository from tracking",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Check if repo exists
		found := false
		for _, repo := range cfg.Repositories {
			if repo.Path == identifier || repo.Name == identifier {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("repository not found: %s", identifier)
		}

		if err := config.RemoveRepo(cfg, identifier); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Removed repository: %s\n", identifier)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}

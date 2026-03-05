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

		cfg, err := config.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		repo := config.GetRepoByName(cfg, identifier)
		if repo == nil {
			repo = config.GetRepoByPath(cfg, identifier)
		}

		if repo == nil {
			return fmt.Errorf("%w: %s", config.ErrRepoNotFound, identifier)
		}

		if err := config.RemoveRepo(cfg, identifier); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Removed repository: %s\n", repo.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}

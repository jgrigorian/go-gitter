package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jgrigorian/go-gitter/internal/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <path> [name]",
	Short: "Add a repository to track",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		// Resolve to absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		// Check if directory exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", absPath)
		}

		// Check if it's a git repository
		gitDir := filepath.Join(absPath, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			return fmt.Errorf("not a git repository: %s", absPath)
		}

		// Determine name
		name := filepath.Base(absPath)
		if len(args) > 1 {
			name = args[1]
		}

		// Get group flag
		group, _ := cmd.Flags().GetString("group")

		// Load config, add repo, save
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := config.AddRepo(cfg, absPath, name, group); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Added repository: %s (%s)\n", name, absPath)
		if group != "" {
			fmt.Printf("Group: %s\n", group)
		}
		return nil
	},
}

func init() {
	addCmd.Flags().StringP("group", "g", "", "Group name for the repository")
	rootCmd.AddCommand(addCmd)
}

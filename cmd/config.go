package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jgrigorian/go-gitter/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.ConfigPath()
		fmt.Println(path)
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open config in editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.ConfigPath()

		// Check if config exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Create default config
			cfg := config.DefaultConfig()
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to create config: %w", err)
			}
			fmt.Printf("Created new config at: %s\n", path)
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		editCmd := exec.Command(editor, path)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr

		return editCmd.Run()
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate config and check repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Println("Validating configuration...")

		if len(cfg.Repositories) == 0 {
			fmt.Println("No repositories configured.")
			return nil
		}

		errors := config.ValidateConfig(cfg)
		if len(errors) == 0 {
			fmt.Printf("✓ Config valid: %d repository(s)\n", len(cfg.Repositories))
			return nil
		}

		fmt.Printf("✗ Found %d issue(s):\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  - %v\n", e)
		}

		return nil
	},
}

func init() {
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configValidateCmd)
	rootCmd.AddCommand(configCmd)
}

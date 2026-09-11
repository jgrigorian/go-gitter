package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func writeCompletion(shell string, output io.Writer) error {
	switch shell {
	case "bash":
		return rootCmd.GenBashCompletion(output)
	case "zsh":
		return rootCmd.GenZshCompletion(output)
	case "fish":
		return rootCmd.GenFishCompletion(output, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletion(output)
	default:
		return fmt.Errorf("unsupported shell %q", shell)
	}
}

var completionCmd = &cobra.Command{
	Use:       "completion [bash|zsh|fish|powershell]",
	Short:     "Generate shell completion script",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return writeCompletion(args[0], cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

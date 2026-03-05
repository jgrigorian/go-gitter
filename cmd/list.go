package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/jgrigorian/go-gitter/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.Repositories) == 0 {
			fmt.Println("No repositories tracked yet.")
			fmt.Println("Run 'go-gitter add <path>' to add a repository.")
			return nil
		}

		group, _ := cmd.Flags().GetString("group")
		showBranch, _ := cmd.Flags().GetBool("branch")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		if showBranch {
			fmt.Fprintln(w, "NAME\tPATH\tGROUP\tBRANCH\tLAST SYNC")
		} else {
			fmt.Fprintln(w, "NAME\tPATH\tGROUP\tLAST SYNC")
		}

		for _, repo := range cfg.Repositories {
			if group != "" && repo.Group != group {
				continue
			}

			lastSync := "-"
			if repo.LastSync != nil {
				lastSync = repo.LastSync.Format("2006-01-02 15:04")
			}

			grp := repo.Group
			if grp == "" {
				grp = "-"
			}

			branch := "-"
			if showBranch {
				branchCmd := exec.Command("git", "branch", "--show-current")
				branchCmd.Dir = repo.Path
				branchOutput, err := branchCmd.Output()
				if err == nil {
					branch = strings.TrimSpace(string(branchOutput))
					if branch == "" {
						branch = "(detached)"
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", repo.Name, repo.Path, grp, branch, lastSync)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", repo.Name, repo.Path, grp, lastSync)
			}
		}
		w.Flush()

		return nil
	},
}

func init() {
	listCmd.Flags().StringP("group", "g", "", "Filter by group")
	listCmd.Flags().BoolP("branch", "b", false, "Show current branch")
	rootCmd.AddCommand(listCmd)
}

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

func getDefaultBranch(path string) string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "main"
	}
	ref := strings.TrimSpace(string(output))
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

func getCommitsBehind(path, defaultBranch string) string {
	cmd := exec.Command("git", "rev-list", "--count", "--left-right", fmt.Sprintf("origin/%s...HEAD", defaultBranch))
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "-"
	}
	parts := strings.Fields(string(output))
	if len(parts) < 2 {
		return "-"
	}
	behind := strings.TrimSpace(parts[0])
	if behind == "0" {
		return "-"
	}
	return behind
}

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

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tPATH\tGROUP\tBRANCH\tBEHIND\tLAST SYNC")

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

			branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
			branchCmd.Dir = repo.Path
			branchOutput, err := branchCmd.Output()
			branch := "-"
			if err == nil {
				branch = strings.TrimSpace(string(branchOutput))
				if branch == "HEAD" {
					branch = "(detached)"
				}
			}

			behind := "-"
			if branch != "-" && branch != "(detached)" {
				defaultBranch := getDefaultBranch(repo.Path)
				behind = getCommitsBehind(repo.Path, defaultBranch)
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", repo.Name, repo.Path, grp, branch, behind, lastSync)
		}
		w.Flush()

		return nil
	},
}

func init() {
	listCmd.Flags().StringP("group", "g", "", "Filter by group")
	rootCmd.AddCommand(listCmd)
}

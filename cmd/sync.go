package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jgrigorian/go-gitter/internal/config"
	"github.com/spf13/cobra"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("blue"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))
	repoStyle    = lipgloss.NewStyle().Bold(true)

	spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

// SyncResult holds the result of syncing a single repository
type SyncResult struct {
	Name              string
	Success           bool
	Warning           bool
	Message           string
	Branch            string
	NonStandardBranch bool
	OriginalIndex     int // Index in original config.Repositories slice
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all tracked repositories (git fetch)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.Repositories) == 0 {
			fmt.Println("No repositories to sync.")
			return nil
		}

		// Get flags
		group, _ := cmd.Flags().GetString("group")
		pull, _ := cmd.Flags().GetBool("pull")

		// Build a slice of repos to sync with their original indices
		type repoWithIndex struct {
			repo    config.Repository
			origIdx int
		}
		reposToSync := make([]repoWithIndex, 0, len(cfg.Repositories))

		for i, r := range cfg.Repositories {
			if group == "" || r.Group == group {
				reposToSync = append(reposToSync, repoWithIndex{repo: r, origIdx: i})
			}
		}

		now := time.Now()

		// Channels and waitgroup for parallel execution
		resultChan := make(chan SyncResult, len(reposToSync))
		var wg sync.WaitGroup

		// Mutex for protecting config updates
		var mu sync.Mutex

		// Print header
		fmt.Println(infoStyle.Render(fmt.Sprintf("Syncing %d repository(s) in parallel...\n", len(reposToSync))))

		// Start spinner animation in background
		done := make(chan bool)
		go func() {
			spinnerIdx := 0
			for {
				select {
				case <-done:
					return
				default:
					fmt.Printf("\r%s %s", infoStyle.Render("⠿"), infoStyle.Render("syncing..."))
					spinnerIdx = (spinnerIdx + 1) % len(spinnerChars)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()

		// Launch goroutines for each repo
		for _, ri := range reposToSync {
			wg.Add(1)
			go func(ri repoWithIndex) {
				defer wg.Done()
				repo := ri.repo

				result := SyncResult{
					Name:          repo.Name,
					OriginalIndex: ri.origIdx,
				}

				// Check if directory exists
				if _, err := os.Stat(repo.Path); os.IsNotExist(err) {
					result.Success = false
					result.Message = "directory not found"
					resultChan <- result
					return
				}

				// Get current branch
				branchCmd := exec.Command("git", "branch", "--show-current")
				branchCmd.Dir = repo.Path
				branchOutput, err := branchCmd.Output()
				result.Branch = strings.TrimSpace(string(branchOutput))

				// Check if branch is non-standard (not main, master, or empty for detached HEAD)
				if err == nil && result.Branch != "" && result.Branch != "main" && result.Branch != "master" {
					result.NonStandardBranch = true
				}

				// Run git fetch with verbose output to detect if there were updates
				gitCmd := exec.Command("git", "fetch", "--all", "--prune")
				gitCmd.Dir = repo.Path
				stderr, err := gitCmd.StderrPipe()
				if err != nil {
					result.Success = false
					result.Message = fmt.Sprintf("failed to run git: %v", err)
					resultChan <- result
					return
				}

				if err := gitCmd.Start(); err != nil {
					result.Success = false
					result.Message = fmt.Sprintf("failed to start git: %v", err)
					resultChan <- result
					return
				}

				stderrOutput, _ := io.ReadAll(stderr)
				if err := gitCmd.Wait(); err != nil {
					result.Success = false
					result.Message = string(stderrOutput)
					if result.Message == "" {
						result.Message = err.Error()
					}
					resultChan <- result
					return
				}

				// Check if there were any updates by comparing refs
				gitRevCmd := exec.Command("git", "rev-list", "--count", "--max-count=1", "HEAD..origin/HEAD")
				gitRevCmd.Dir = repo.Path
				revOutput, _ := gitRevCmd.Output()
				updatesCount := strings.TrimSpace(string(revOutput))

				hasUpdates := updatesCount != "0" && updatesCount != ""

				// If pull flag is set, also do git pull from current branch
				if pull {
					var pullCmd *exec.Cmd
					if result.Branch != "" {
						pullCmd = exec.Command("git", "pull", "origin", result.Branch)
					} else {
						pullCmd = exec.Command("git", "pull")
					}
					pullCmd.Dir = repo.Path
					if output, err := pullCmd.CombinedOutput(); err != nil {
						result.Warning = true
						result.Message = fmt.Sprintf("pull failed: %s", string(output))
					} else {
						result.Success = true
						result.Message = "pulled latest"
					}
				} else {
					result.Success = true
					if hasUpdates {
						result.Message = fmt.Sprintf("fetched %s update(s)", updatesCount)
					} else {
						result.Message = "already up to date"
					}
				}

				// Update last sync time (thread-safe) using original index
				if result.Success || result.Warning {
					mu.Lock()
					cfg.Repositories[ri.origIdx].LastSync = &now
					mu.Unlock()
				}

				resultChan <- result
			}(ri)
		}

		// Wait for all goroutines and collect results
		results := make([]SyncResult, 0, len(reposToSync))
		for range reposToSync {
			result := <-resultChan
			results = append(results, result)
		}

		// Stop spinner
		close(done)

		// Clear spinner line
		fmt.Printf("\r%s\r", "\033[K")

		// Print results
		successCount := 0
		failCount := 0

		for _, result := range results {
			branchInfo := ""
			if result.Branch != "" {
				branchInfo = fmt.Sprintf("(%s) ", result.Branch)
			}

			if result.Success {
				if result.NonStandardBranch {
					fmt.Printf("%s %s %s%s %s\n", warnStyle.Render("⚠"), repoStyle.Render(result.Name), branchInfo, warnStyle.Render(result.Message), warnStyle.Render("branch: "+result.Branch))
				} else {
					fmt.Printf("%s %s %s%s\n", successStyle.Render("✓"), repoStyle.Render(result.Name), branchInfo, successStyle.Render(result.Message))
				}
				successCount++
			} else if result.Warning {
				fmt.Printf("%s %s %s%s\n", warnStyle.Render("⚠"), repoStyle.Render(result.Name), branchInfo, warnStyle.Render(result.Message))
				successCount++
			} else {
				fmt.Printf("%s %s %s\n", successStyle.Render("✗"), repoStyle.Render(result.Name), errorStyle.Render(result.Message))
				failCount++
			}
		}

		// Save updated config with last sync times
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		// Print summary
		fmt.Println()
		if failCount > 0 {
			fmt.Printf("%s %s, %s\n",
				successStyle.Render(fmt.Sprintf("%d succeeded", successCount)),
				errorStyle.Render(fmt.Sprintf("%d failed", failCount)),
				infoStyle.Render(time.Now().Format("15:04:05")))
		} else {
			fmt.Printf("%s %s\n",
				successStyle.Render(fmt.Sprintf("All %d repositories synced!", successCount)),
				infoStyle.Render(time.Now().Format("15:04:05")))
		}

		return nil
	},
}

func init() {
	syncCmd.Flags().StringP("group", "g", "", "Sync only repositories in this group")
	syncCmd.Flags().BoolP("pull", "p", false, "Also pull (not just fetch)")
	rootCmd.AddCommand(syncCmd)
}

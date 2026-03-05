package cmd

import (
	"fmt"
	"os"
	"os/exec"
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
	Name      string
	Success   bool
	Warning   bool
	Message   string
	RepoIndex int // Index in original repos slice
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

		// Filter by group if specified
		group, _ := cmd.Flags().GetString("group")
		pull, _ := cmd.Flags().GetBool("pull")

		repos := cfg.Repositories
		if group != "" {
			var filtered []config.Repository
			for _, r := range repos {
				if r.Group == group {
					filtered = append(filtered, r)
				}
			}
			repos = filtered
		}

		now := time.Now()

		// Channels and waitgroup for parallel execution
		resultChan := make(chan SyncResult, len(repos))
		var wg sync.WaitGroup

		// Mutex for protecting config updates
		var mu sync.Mutex

		// Print header
		fmt.Println(infoStyle.Render(fmt.Sprintf("Syncing %d repository(s) in parallel...\n", len(repos))))

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
		for i := range repos {
			wg.Add(1)
			go func(i int, repo config.Repository) {
				defer wg.Done()

				result := SyncResult{
					Name:      repo.Name,
					RepoIndex: i,
				}

				// Check if directory exists
				if _, err := os.Stat(repo.Path); os.IsNotExist(err) {
					result.Success = false
					result.Message = "directory not found"
					resultChan <- result
					return
				}

				// Run git fetch
				gitCmd := exec.Command("git", "fetch", "--all")
				gitCmd.Dir = repo.Path
				if output, err := gitCmd.CombinedOutput(); err != nil {
					result.Success = false
					result.Message = string(output)
					resultChan <- result
					return
				}

				// If pull flag is set, also do git pull
				if pull {
					pullCmd := exec.Command("git", "pull")
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
					result.Message = "fetched updates"
				}

				// Update last sync time (thread-safe)
				if result.Success || result.Warning {
					mu.Lock()
					cfg.Repositories[i].LastSync = &now
					mu.Unlock()
				}

				resultChan <- result
			}(i, repos[i])
		}

		// Wait for all goroutines and collect results
		results := make([]SyncResult, 0, len(repos))
		for range repos {
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
			if result.Success {
				fmt.Printf("%s %s %s\n", successStyle.Render("✓"), repoStyle.Render(result.Name), successStyle.Render(result.Message))
				successCount++
			} else if result.Warning {
				fmt.Printf("%s %s %s\n", warnStyle.Render("⚠"), repoStyle.Render(result.Name), warnStyle.Render(result.Message))
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

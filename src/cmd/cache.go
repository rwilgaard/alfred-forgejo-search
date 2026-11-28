package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "refresh repository cache",
	RunE: func(_ *cobra.Command, _ []string) error {
		log.Println("[cache] fetching repositories...")
		if err := fetchAndCacheRepos(); err != nil {
			// Background job: Alfred isn't listening, so report through the
			// error trigger instead of returning to FatalError.
			wf.ReportError(fmt.Errorf("cache failed: %w", err))
			return nil
		}
		log.Println("[cache] repositories fetched")
		return nil
	},
}

func fetchAndCacheRepos() error {
	client, err := setupClient(nil)
	if err != nil {
		return fmt.Errorf("error retrieving credentials: %w", err)
	}
	repos, err := client.GetAllRepos()
	if err != nil {
		return fmt.Errorf("could not fetch repos: %w", err)
	}
	if err := wf.Cache.StoreJSON(repoCacheName, repos); err != nil {
		return fmt.Errorf("could not write cache file: %w", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}

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
			return err
		}
		log.Println("[cache] repositories fetched")
		return nil
	},
}

func fetchAndCacheRepos() error {
	client, err := setupClient()
	if err != nil {
		msg := fmt.Sprintf("Cache failed: error retrieving credentials: %s", err)
		if zerr := wf.Alfred.RunTrigger("error", msg); zerr != nil {
			log.Printf("Alfred error trigger failed: %v", zerr)
		}
		return err
	}

	repos, err := client.GetAllRepos()
	if err != nil {
		msg := fmt.Sprintf("Cache failed: could not fetch repos: %s", err)
		if zerr := wf.Alfred.RunTrigger("error", msg); zerr != nil {
			log.Printf("Alfred error trigger failed: %v", zerr)
		}
		return err
	}

	if err := wf.Cache.StoreJSON(repoCacheName, repos); err != nil {
		msg := fmt.Sprintf("Cache failed: could not write cache file: %s", err)
		if zerr := wf.Alfred.RunTrigger("error", msg); zerr != nil {
			log.Printf("Alfred error trigger failed: %v", zerr)
		}
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}

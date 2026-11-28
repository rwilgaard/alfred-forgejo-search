package cmd

import (
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/rwilgaard/go-alfredutils/v2/alfred"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [query]",
	Short: "list user repositories",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if !wf.RequireAuth(keychainAccount) {
			return
		}

		var repos []*gitea.Repository
		if err := wf.Cache.LoadJSONIfExists(repoCacheName, &repos); err != nil {
			wf.FatalError(err)
		}
		maxAge := time.Duration(cfg.CacheAge) * time.Minute
		if _, err := wf.RefreshIfStale(repoCacheName, maxAge, "cache"); err != nil {
			wf.FatalError(err)
		}

		var query string
		if len(args) > 0 {
			query = args[0]
		}

		for _, r := range repos {
			// ⏎ opens the repository in the browser; ⌘⏎ opens the details
			// window, which reads these variables. Set the variables before
			// NewModifier, which copies them.
			it := wf.NewItem(r.Name).
				UID(r.FullName).
				Subtitle(buildRepoSubtitle(r)).
				MatchOn(alfred.MatchSecondary, r.FullName).
				MatchOn(alfred.MatchMeta, r.Description).
				Var("repo_fullname", r.FullName).
				Var("repo_html_url", r.HTMLURL).
				Var("repo_clone_url", r.CloneURL).
				Var("repo_ssh_url", r.SSHURL).
				Arg("open").
				Valid(true)
			it.NewModifier(alfred.ModCmd).
				Subtitle("Show details, issues and pull requests").
				Arg("details").
				Valid(true)
		}

		wf.Filter(query)
		wf.Send()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

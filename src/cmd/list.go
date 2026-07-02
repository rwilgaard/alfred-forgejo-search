package cmd

import (
	"fmt"
	"time"

	"code.gitea.io/sdk/gitea"
	aw "github.com/deanishe/awgo"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [query]",
	Short: "list user repositories",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if ok := alfredutils.HandleAuthentication(wf, keychainAccount); !ok {
			return
		}

		var repos []*gitea.Repository
		if err := alfredutils.LoadCache(wf, repoCacheName, &repos); err != nil {
			wf.FatalError(err)
		}

		maxCacheAge := time.Duration(cfg.CacheAge) * time.Minute
		if err := alfredutils.RefreshCache(wf, repoCacheName, maxCacheAge, []string{"cache"}); err != nil {
			wf.FatalError(err)
		}

		var query string
		if len(args) > 0 {
			query = args[0]
		}

		for _, r := range repos {
			subtitle := buildRepoSubtitle(r)
			item := wf.NewItem(r.Name).
				UID(r.FullName).
				Subtitle(subtitle).
				Var("item_url", r.HTMLURL).
				Arg("open").
				Valid(true)

			item.NewModifier(aw.ModCmd).
				Subtitle(fmt.Sprintf("Copy HTTPS clone URL: %s", r.CloneURL)).
				Arg(r.CloneURL).
				Valid(true)

			item.NewModifier(aw.ModOpt).
				Subtitle(fmt.Sprintf("Copy SSH clone URL: %s", r.SSHURL)).
				Arg(r.SSHURL).
				Valid(true)

			item.NewModifier(aw.ModCtrl).
				Subtitle(fmt.Sprintf("Show issues for %s", r.FullName)).
				Var("repo_fullname", r.FullName).
				Var("list_query", query).
				Arg("").
				Valid(true)

			item.NewModifier(aw.ModShift).
				Subtitle(fmt.Sprintf("Show pull requests for %s", r.FullName)).
				Var("repo_fullname", r.FullName).
				Var("list_query", query).
				Arg("").
				Valid(true)
		}

		if query != "" {
			wf.Filter(query)
		}

		alfredutils.HandleFeedback(wf)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

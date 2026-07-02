package cmd

import (
	"fmt"
	"strings"

	aw "github.com/deanishe/awgo"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var issuesCmd = &cobra.Command{
	Use:   "issues [filter]",
	Short: "list open issues for the repository in repo_fullname env var",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if ok := alfredutils.HandleAuthentication(wf, keychainAccount); !ok {
			return
		}

		repoFullName := cfg.RepoFullName
		if repoFullName == "" {
			wf.NewItem("No repository selected").
				Subtitle("Select a repository first with ⌃⏎").
				Icon(aw.IconWarning).
				Valid(false)
			alfredutils.HandleFeedback(wf)
			return
		}

		parts := strings.SplitN(repoFullName, "/", 2)
		if len(parts) != 2 {
			wf.NewItem("Invalid repository").
				Subtitle(repoFullName).
				Icon(aw.IconError).
				Valid(false)
			alfredutils.HandleFeedback(wf)
			return
		}
		owner, repoName := parts[0], parts[1]

		client, err := setupClient()
		if err != nil {
			wf.FatalError(err)
		}

		issues, err := client.GetIssues(owner, repoName)
		if err != nil {
			wf.FatalError(err)
		}

		wf.NewItem(fmt.Sprintf("← %s", repoFullName)).
			Subtitle("Back to repositories").
			Icon(aw.IconHome).
			Var("list_query", cfg.ListQuery).
			Arg("back").
			Valid(true)

		for _, issue := range issues {
			title := fmt.Sprintf("#%d  %s", issue.Index, issue.Title)
			subtitle := buildIssueSubtitle(issue)
			wf.NewItem(title).
				Subtitle(subtitle).
				Var("item_url", issue.HTMLURL).
				Arg("open").
				Valid(true)
		}

		var filter string
		if len(args) > 0 {
			filter = args[0]
		}
		if filter != "" {
			wf.Filter(filter)
		}

		alfredutils.HandleFeedback(wf)
	},
}

func init() {
	rootCmd.AddCommand(issuesCmd)
}

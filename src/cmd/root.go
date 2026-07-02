package cmd

import (
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"github.com/maniartech/gotime"
	forgejo "github.com/rwilgaard/alfred-forgejo-search/src/internal/forgejo"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
	"go.deanishe.net/fuzzy"
)

type workflowConfig struct {
	URL          string `env:"forgejo_url"`
	CacheAge     int    `env:"cache_age"`
	RepoFullName string `env:"repo_fullname"`
	ListQuery    string `env:"list_query"`
}

const (
	repo            = "rwilgaard/alfred-forgejo-search"
	repoCacheName   = "repositories.json"
	keychainAccount = "alfred-forgejo-search"
)

var (
	wf      *aw.Workflow
	cfg     = &workflowConfig{}
	rootCmd = &cobra.Command{
		Use:           "alfred-forgejo-search",
		Short:         "Alfred workflow for searching Forgejo repositories",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

// Execute is the entrypoint called from main.
func Execute() {
	wf.Run(run)
}

func run() {
	alfredutils.AddClearAuthMagic(wf, keychainAccount)

	if err := alfredutils.InitWorkflow(wf, cfg); err != nil {
		wf.FatalError(err)
	}
	if err := alfredutils.CheckForUpdates(wf); err != nil {
		wf.FatalError(err)
	}
	if err := rootCmd.Execute(); err != nil {
		wf.FatalError(err)
	}
}

func setupClient() (*forgejo.Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("forgejo_url is not configured in workflow settings")
	}
	token, err := wf.Keychain.Get(keychainAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get token from keychain: %w", err)
	}
	return forgejo.NewClient(cfg.URL, token)
}

func buildRepoSubtitle(r *gitea.Repository) string {
	var lastPush string
	if !r.Updated.IsZero() {
		lastPush = gotime.TimeAgo(r.Updated)
	} else {
		lastPush = "never"
	}

	parts := []string{
		r.Owner.UserName,
		fmt.Sprintf("★ %d", r.Stars),
		lastPush,
	}
	if r.Description != "" {
		parts = append(parts, r.Description)
	}
	return strings.Join(parts, "  ·  ")
}

func buildIssueSubtitle(issue *gitea.Issue) string {
	parts := []string{
		fmt.Sprintf("#%d", issue.Index),
		gotime.TimeAgo(issue.Created),
		fmt.Sprintf("%d comments", issue.Comments),
	}
	if issue.Poster != nil {
		parts = append([]string{issue.Poster.UserName}, parts...)
	}
	return strings.Join(parts, "  ·  ")
}

func buildPRSubtitle(pr *gitea.PullRequest) string {
	head := ""
	base := ""
	if pr.Head != nil {
		head = pr.Head.Ref
	}
	if pr.Base != nil {
		base = pr.Base.Ref
	}
	parts := []string{
		fmt.Sprintf("#%d", pr.Index),
		fmt.Sprintf("%s → %s", head, base),
		fmt.Sprintf("%d comments", pr.Comments),
	}
	if pr.Draft {
		parts = append([]string{"[Draft]"}, parts...)
	}
	return strings.Join(parts, "  ·  ")
}

func init() {
	sopts := []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	wf = aw.New(
		aw.SortOptions(sopts...),
		update.GitHub(repo),
	)
}

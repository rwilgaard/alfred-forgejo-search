package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/maniartech/gotime/v2"
	forgejo "github.com/rwilgaard/alfred-forgejo-search/src/internal/forgejo"
	"github.com/rwilgaard/go-alfredutils/v2/alfred"
	"github.com/rwilgaard/go-alfredutils/v2/alfred/update"
	"github.com/spf13/cobra"
	"go.deanishe.net/fuzzy"
)

type workflowConfig struct {
	URL      string `env:"forgejo_url"`
	CacheAge int    `env:"cache_age"`

	// Set by the list item that opens the repository window.
	RepoFullName string `env:"repo_fullname"`
	RepoCloneURL string `env:"repo_clone_url"`
	RepoSSHURL   string `env:"repo_ssh_url"`
}

const (
	repo            = "rwilgaard/alfred-forgejo-search"
	repoCacheName   = "repositories.json"
	keychainAccount = "alfred-forgejo-search"
)

var (
	wf      *alfred.Workflow
	updater *update.Updater
	cfg     = &workflowConfig{CacheAge: 360}
	rootCmd = &cobra.Command{
		Use:           "alfred-forgejo-search",
		Short:         "Alfred workflow for searching Forgejo repositories",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

// Execute is the entrypoint called from main.
func Execute() {
	var err error
	wf, err = alfred.New(
		alfred.WithUsage(),
		alfred.WithFuzzy(
			fuzzy.AdjacencyBonus(10.0),
			fuzzy.LeadingLetterPenalty(-0.1),
			fuzzy.MaxLeadingLetterPenalty(-3.0),
			fuzzy.UnmatchedLetterPenalty(-0.5),
		),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Both register "workflow:" magic queries, which wf.Run handles before
	// run(), so they must come first.
	updater = update.GitHub(wf, repo)
	wf.AddClearAuth(keychainAccount)
	wf.Run(run)
}

func run() {
	if err := wf.Config(cfg); err != nil {
		wf.FatalError(err)
	}
	if err := updater.Run(); err != nil {
		wf.FatalError(err)
	}
	if err := rootCmd.Execute(); err != nil {
		wf.FatalError(err)
	}
}

// setupClient returns an authenticated client. A nil httpClient uses the
// SDK's default client.
func setupClient(httpClient *http.Client) (*forgejo.Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("forgejo_url is not configured in workflow settings")
	}
	token, err := wf.LoadAuth(keychainAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get token from keychain: %w", err)
	}
	return forgejo.NewClient(cfg.URL, token, httpClient)
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

package cmd

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/maniartech/gotime/v2"
	forgejo "github.com/rwilgaard/alfred-forgejo-search/src/internal/forgejo"
	"github.com/rwilgaard/alfred-forgejo-search/src/internal/preview"
	"github.com/rwilgaard/go-alfredutils/v2/alfred/clipboard"
	"github.com/rwilgaard/go-alfredutils/v2/alfred/httpcache"
	alfredui "github.com/rwilgaard/go-alfredutils/v2/alfred/ui"
	"github.com/spf13/cobra"
)

// Component ids in workflow/repo.alfredui/ui.json.
const (
	uiTitle    = "title"
	uiSubtitle = "subtitle"
	uiMeta     = "meta"
	uiItems    = "items"
	uiFilter   = "filter"
	uiLoading  = "loading"
	uiIssues   = "issuesbtn"
	uiPulls    = "pullsbtn"
)

// Arguments of ui-repo-items.
const (
	kindIssues = "issues"
	kindPulls  = "pulls"
)

// repoCacheAge is how long issue, pull request and repository responses are
// reused, so switching lists or reopening the window doesn't wait on Forgejo.
const repoCacheAge = time.Minute

// uiRepoCmd fills in the repository window when it opens: the header and the
// open issues.
var uiRepoCmd = &cobra.Command{
	Use:    "ui-repo",
	Short:  "fill in the repository window",
	Hidden: true,
	Run: func(_ *cobra.Command, _ []string) {
		emit(uiRepo(cfg.RepoFullName))
	},
}

// uiRepoItemsCmd switches the list. The window fetched both lists on open, so
// this is answered from the HTTP cache.
var uiRepoItemsCmd = &cobra.Command{
	Use:       "ui-repo-items issues|pulls",
	Short:     "show issues or pull requests in the repository window",
	Hidden:    true,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{kindIssues, kindPulls},
	Run: func(_ *cobra.Command, args []string) {
		emit(uiRepoItems(cfg.RepoFullName, args[0]))
	},
}

// uiCopyCmd copies a clone URL and keeps the window open.
var uiCopyCmd = &cobra.Command{
	Use:       "ui-copy https|ssh",
	Short:     "copy a clone URL from the repository window",
	Hidden:    true,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"https", "ssh"},
	Run: func(_ *cobra.Command, args []string) {
		emit(uiCopy(args[0]))
	},
}

func repoClient() (*forgejo.Client, error) {
	return setupClient(httpcache.New(wf.Cache, repoCacheAge))
}

func splitRepoName(fullName string) (owner, name string, err error) {
	owner, name, ok := strings.Cut(strings.TrimSpace(fullName), "/")
	if !ok || owner == "" || name == "" {
		return "", "", fmt.Errorf("invalid repository %q", fullName)
	}
	return owner, name, nil
}

func uiRepo(fullName string) alfredui.Actions {
	a := alfredui.NewUpdate()
	owner, name, err := splitRepoName(fullName)
	if err != nil {
		return failLoading(a, err)
	}
	client, err := repoClient()
	if err != nil {
		return failLoading(a, err)
	}

	var (
		wg         sync.WaitGroup
		repo       *gitea.Repository
		rows       []alfredui.Props
		rErr, lErr error
	)
	wg.Go(func() { repo, rErr = client.GetRepo(owner, name) })
	wg.Go(func() { rows, lErr = fetchRows(client, owner, name, kindIssues) })
	// Fetched only to fill the HTTP cache, so switching to pull requests
	// doesn't wait.
	wg.Go(func() {
		if _, err := fetchRows(client, owner, name, kindPulls); err != nil {
			log.Printf("[ui] prefetch pull requests: %v", err)
		}
	})
	wg.Wait()

	if rErr != nil {
		return failLoading(a, rErr)
	}
	if err := wf.Usage.Record(fullName); err != nil {
		log.Printf("[ui] record usage of %s: %v", fullName, err)
	}

	a = a.Set(uiTitle, alfredui.Props{"text": repo.FullName}).
		Set(uiSubtitle, alfredui.Props{"text": orSpace(repo.Description)}).
		Set(uiMeta, alfredui.Props{"text": repoMeta(repo)})
	a.Update.Relayout = true // the header labels get their width from this text

	if lErr != nil {
		return failLoading(a, lErr)
	}
	return showRows(a, kindIssues, rows)
}

func uiRepoItems(fullName, kind string) alfredui.Actions {
	a := alfredui.NewUpdate()
	owner, name, err := splitRepoName(fullName)
	if err != nil {
		return failLoading(a, err)
	}
	client, err := repoClient()
	if err != nil {
		return failLoading(a, err)
	}
	if kind != kindPulls {
		kind = kindIssues
	}
	rows, err := fetchRows(client, owner, name, kind)
	if err != nil {
		return failLoading(a, err)
	}
	return showRows(a, kind, rows)
}

func fetchRows(client *forgejo.Client, owner, name, kind string) ([]alfredui.Props, error) {
	if kind == kindPulls {
		prs, err := client.GetPullRequests(owner, name)
		if err != nil {
			return nil, err
		}
		return pullRows(prs), nil
	}
	issues, err := client.GetIssues(owner, name)
	if err != nil {
		return nil, err
	}
	return issueRows(issues), nil
}

func uiCopy(which string) alfredui.Actions {
	value, label := cfg.RepoCloneURL, "HTTPS URL"
	if which == "ssh" {
		value, label = cfg.RepoSSHURL, "SSH URL"
	}
	if value == "" {
		return status("No "+label+" for this repository", alfredui.ColourError)
	}
	if _, err := clipboard.Copy(value); err != nil {
		return status(err.Error(), alfredui.ColourError)
	}
	return status("✓ Copied "+label, alfredui.ColourSuccess)
}

func failLoading(a alfredui.Actions, err error) alfredui.Actions {
	log.Printf("[ui] %v", err)
	a = a.Set(uiLoading, alfredui.Props{"hidden": true}).
		Status(uiStatus, err.Error(), alfredui.ColourError)
	a.Update.Relayout = true // see status
	return a
}

// showRows puts rows in the list and marks which list it is: the active
// toggle button is bold, and the filter's placeholder names the list and its
// count. The filter is sent with the data
// so it survives the update.
func showRows(a alfredui.Actions, kind string, rows []alfredui.Props) alfredui.Actions {
	on, off := uiIssues, uiPulls
	noun := "issues"
	if kind == kindPulls {
		on, off = uiPulls, uiIssues
		noun = "pull requests"
	}

	loading := alfredui.Props{"hidden": true}
	if len(rows) == 0 {
		loading = alfredui.Props{"hidden": false, "text": "No open " + noun}
	}
	placeholder := fmt.Sprintf("Filter %s open %s   ·   ↩ open selected",
		countText(len(rows), forgejo.PageSize), noun)

	return a.Set(uiItems, alfredui.Props{"data": map[string]any{
		"type": "basic",
		"src":  rows,
		"filter": map[string]any{
			"type":  "basic",
			"args":  []string{"rowtitle.text", "rowsubtitle.text"},
			"match": "{value:filter.text}",
		},
	}}).
		Set(on, alfredui.Props{"font": map[string]any{"weight": "bold"}}).
		Set(off, alfredui.Props{"font": nil}).
		Set(uiFilter, alfredui.Props{"placeholder": placeholder}).
		Set(uiLoading, loading).
		Status(uiStatus, " ", alfredui.ColourInfo)
}

func issueRows(issues []*gitea.Issue) []alfredui.Props {
	rows := make([]alfredui.Props, 0, len(issues))
	for _, iss := range issues {
		author := userName(iss.Poster)
		labels := labelNames(iss.Labels)
		parts := []string{fmt.Sprintf("#%d", iss.Index), "by " + author, timeAgo(&iss.Updated)}
		if c := comments(iss.Comments); c != "" {
			parts = append(parts, c)
		}
		if labels != "" {
			parts = append(parts, labels)
		}

		detail := []string{
			"Open issue",
			fmt.Sprintf("opened %s by %s", lowerTime(&iss.Created), author),
			"updated " + lowerTime(&iss.Updated),
		}
		if c := comments(iss.Comments); c != "" {
			detail = append(detail, c)
		}

		rows = append(rows, row(iss.Title, strings.Join(parts, "  ·  "), iss.HTMLURL, rowPreview{
			title:  fmt.Sprintf("#%d  %s", iss.Index, iss.Title),
			detail: strings.Join(detail, "  ·  "),
			tags:   tagLine(labels, iss.Assignees),
			body:   iss.Body,
		}))
	}
	return rows
}

func pullRows(prs []*gitea.PullRequest) []alfredui.Props {
	rows := make([]alfredui.Props, 0, len(prs))
	for _, pr := range prs {
		author := userName(pr.Poster)
		head, base := branchRef(pr.Head), branchRef(pr.Base)
		parts := []string{fmt.Sprintf("#%d", pr.Index), "by " + author, timeAgo(pr.Updated), head}
		state := "Open pull request"
		if pr.Draft {
			parts = append(parts, "Draft")
			state = "Draft pull request"
		}
		detail := []string{
			state,
			head + " → " + base,
			fmt.Sprintf("opened %s by %s", lowerTime(pr.Created), author),
			"updated " + lowerTime(pr.Updated),
		}

		rows = append(rows, row(pr.Title, strings.Join(parts, "  ·  "), pr.HTMLURL, rowPreview{
			title:  fmt.Sprintf("#%d  %s", pr.Index, pr.Title),
			detail: strings.Join(detail, "  ·  "),
			tags:   tagLine(labelNames(pr.Labels), pr.Assignees),
			body:   pr.Body,
		}))
	}
	return rows
}

// rowPreview is what the ⌘Y sheet shows for a row. It travels in hidden cell
// labels, so opening the sheet needs no request and no process.
type rowPreview struct {
	title, detail, tags, body string
}

func row(title, subtitle, url string, p rowPreview) alfredui.Props {
	return alfredui.Props{
		"rowtitle.text":    title,
		"rowsubtitle.text": subtitle,
		"url.text":         url,
		"ptitle.text":      p.title,
		"pdetail.text":     p.detail,
		"ptags.text":       orSpace(p.tags),
		"pbody.text":       preview.Body(p.body),
	}
}

func tagLine(labels string, assignees []*gitea.User) string {
	var parts []string
	if labels != "" {
		parts = append(parts, "Labels: "+labels)
	}
	if len(assignees) > 0 {
		names := make([]string, 0, len(assignees))
		for _, u := range assignees {
			names = append(names, userName(u))
		}
		parts = append(parts, "Assignees: "+strings.Join(names, ", "))
	}
	return strings.Join(parts, "  ·  ")
}

func labelNames(labels []*gitea.Label) string {
	names := make([]string, 0, len(labels))
	for _, l := range labels {
		names = append(names, l.Name)
	}
	return strings.Join(names, ", ")
}

func repoMeta(r *gitea.Repository) string {
	parts := []string{"Public"}
	if r.Private {
		parts[0] = "Private"
	}
	if r.Archived {
		parts = append(parts, "Archived")
	}
	if r.Fork {
		parts = append(parts, "Fork")
	}
	if r.Language != "" {
		parts = append(parts, r.Language)
	}
	parts = append(parts,
		fmt.Sprintf("★ %d", r.Stars),
		fmt.Sprintf("%d forks", r.Forks),
		"updated "+lowerTime(&r.Updated),
	)
	if len(r.Topics) > 0 {
		parts = append(parts, strings.Join(r.Topics, ", "))
	}
	return strings.Join(parts, "  ·  ")
}

func userName(u *gitea.User) string {
	if u == nil || u.UserName == "" {
		return "unknown"
	}
	return u.UserName
}

func branchRef(b *gitea.PRBranchInfo) string {
	if b == nil {
		return "?"
	}
	return b.Ref
}

func comments(n int) string {
	switch n {
	case 0:
		return ""
	case 1:
		return "1 comment"
	default:
		return fmt.Sprintf("%d comments", n)
	}
}

func timeAgo(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "never"
	}
	return gotime.TimeAgo(*t)
}

func lowerTime(t *time.Time) string { return strings.ToLower(timeAgo(t)) }

// countText shows "50+" when a list hit the page limit.
func countText(n, limit int) string {
	if n >= limit {
		return fmt.Sprintf("%d+", limit)
	}
	return fmt.Sprint(n)
}

func init() {
	rootCmd.AddCommand(uiRepoCmd, uiRepoItemsCmd, uiCopyCmd)
}

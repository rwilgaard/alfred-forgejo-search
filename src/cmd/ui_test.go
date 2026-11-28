package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"code.gitea.io/sdk/gitea"
	alfredui "github.com/rwilgaard/go-alfredutils/v2/alfred/ui"
)

// render returns the JSON a UI command would print.
func render(t *testing.T, a alfredui.Actions) string {
	t.Helper()
	var buf bytes.Buffer
	if err := alfredui.Write(&buf, a); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func assertContains(t *testing.T, out string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("output missing %q:\n%s", w, out)
		}
	}
}

func TestSplitRepoName(t *testing.T) {
	tests := []struct {
		in          string
		owner, name string
		wantErr     bool
	}{
		{"rw/alfred-forgejo-search", "rw", "alfred-forgejo-search", false},
		{" rw/x ", "rw", "x", false},
		{"", "", "", true},
		{"bad", "", "", true},
		{"/b", "", "", true},
		{"a/", "", "", true},
	}
	for _, tt := range tests {
		owner, name, err := splitRepoName(tt.in)
		if (err != nil) != tt.wantErr || owner != tt.owner || name != tt.name {
			t.Errorf("splitRepoName(%q) = %q, %q, %v", tt.in, owner, name, err)
		}
	}
}

func TestShowRows(t *testing.T) {
	out := render(t, showRows(alfredui.NewUpdate(), kindPulls, nil))
	assertContains(t, out, "No open pull requests", "Filter 0 open pull requests")

	rows := make([]alfredui.Props, 50)
	out = render(t, showRows(alfredui.NewUpdate(), kindIssues, rows))
	assertContains(t, out, "Filter 50+ open issues", `"rowtitle.text"`)
}

func TestRowsSurviveMissingFields(t *testing.T) {
	// Deleted users and PRs without branch info come back as nil.
	issues := issueRows([]*gitea.Issue{{Index: 3, Title: "Bug", HTMLURL: "https://g/i/3"}})
	if got := issues[0]["rowsubtitle.text"]; !strings.Contains(got.(string), "by unknown") {
		t.Errorf("issue subtitle = %q", got)
	}
	prs := pullRows([]*gitea.PullRequest{{Index: 4, Title: "Fix", Draft: true}})
	if got := prs[0]["pdetail.text"]; !strings.HasPrefix(got.(string), "Draft pull request  ·  ? → ?") {
		t.Errorf("pull detail = %q", got)
	}
}

func TestUIAuthRejectsBeforeStoring(t *testing.T) {
	called := false
	out := render(t, uiAuth("  \n", func(string) error { called = true; return nil }))
	if called {
		t.Error("empty token was checked")
	}
	assertContains(t, out, "Enter an access token.", `"enabled":true`)

	out = render(t, uiAuth("tok", func(string) error { return errors.New("authentication failed: 401") }))
	assertContains(t, out, "authentication failed: 401", `"text":"Save"`)
	if strings.Contains(out, "finish") {
		t.Errorf("failed check finished the form:\n%s", out)
	}
}

func TestUIRetry(t *testing.T) {
	if out := render(t, uiRetry(func() error { return nil })); !strings.Contains(out, "finish") {
		t.Errorf("successful retry didn't finish:\n%s", out)
	}
	out := render(t, uiRetry(func() error { return errors.New("dial tcp: refused") }))
	assertContains(t, out, "dial tcp: refused", "Still failing", `"error_msg":"dial tcp: refused"`)
}

func TestUIOpenNeedsURL(t *testing.T) {
	out := render(t, uiOpen("{value:items.selected.url.text}", true))
	assertContains(t, out, "Nothing selected to open", `"close":"preview"`)
	if strings.Contains(out, "finish") {
		t.Errorf("open without URL finished:\n%s", out)
	}
}

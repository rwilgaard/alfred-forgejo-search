package forgejo

import (
	"fmt"
	"net/http"

	"code.gitea.io/sdk/gitea"
)

// PageSize is the number of items asked for per request; issue and pull
// request lists show only the first page.
const PageSize = 50

// Client wraps the Gitea/Forgejo SDK client.
type Client struct {
	c   *gitea.Client
	URL string
}

// NewClient creates an authenticated Forgejo client. A nil httpClient uses
// the SDK's default client.
func NewClient(url, token string, httpClient *http.Client) (*Client, error) {
	opts := []gitea.ClientOption{gitea.SetToken(token)}
	if httpClient != nil {
		opts = append(opts, gitea.SetHTTPClient(httpClient))
	}
	c, err := gitea.NewClient(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create forgejo client: %w", err)
	}
	return &Client{c: c, URL: url}, nil
}

// TestAuthentication verifies the token is valid.
func (cl *Client) TestAuthentication() error {
	_, _, err := cl.c.GetMyUserInfo()
	return err
}

// GetRepo returns one repository.
func (cl *Client) GetRepo(owner, repoName string) (*gitea.Repository, error) {
	r, _, err := cl.c.GetRepo(owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s/%s: %w", owner, repoName, err)
	}
	return r, nil
}

// GetAllRepos returns all repositories accessible to the authenticated user.
func (cl *Client) GetAllRepos() ([]*gitea.Repository, error) {
	var all []*gitea.Repository
	opts := gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{Page: 1, PageSize: PageSize},
	}
	for {
		repos, _, err := cl.c.ListMyRepos(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repos (page %d): %w", opts.Page, err)
		}
		all = append(all, repos...)
		if len(repos) < PageSize {
			break
		}
		opts.Page++
	}
	return all, nil
}

// GetIssues returns open issues for a repository.
func (cl *Client) GetIssues(owner, repoName string) ([]*gitea.Issue, error) {
	issues, _, err := cl.c.ListRepoIssues(owner, repoName, gitea.ListIssueOption{
		Type:        gitea.IssueTypeIssue,
		State:       gitea.StateOpen,
		ListOptions: gitea.ListOptions{PageSize: PageSize},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list issues for %s/%s: %w", owner, repoName, err)
	}
	return issues, nil
}

// GetPullRequests returns open pull requests for a repository.
func (cl *Client) GetPullRequests(owner, repoName string) ([]*gitea.PullRequest, error) {
	prs, _, err := cl.c.ListRepoPullRequests(owner, repoName, gitea.ListPullRequestsOptions{
		State:       gitea.StateOpen,
		ListOptions: gitea.ListOptions{PageSize: PageSize},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs for %s/%s: %w", owner, repoName, err)
	}
	return prs, nil
}

package forgejo

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
)

// Client wraps the Gitea/Forgejo SDK client.
type Client struct {
	c   *gitea.Client
	URL string
}

// NewClient creates an authenticated Forgejo client.
func NewClient(url, token string) (*Client, error) {
	c, err := gitea.NewClient(url, gitea.SetToken(token))
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

// GetAllRepos returns all repositories accessible to the authenticated user.
func (cl *Client) GetAllRepos() ([]*gitea.Repository, error) {
	var all []*gitea.Repository
	const pageSize = 50
	opts := gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{Page: 1, PageSize: pageSize},
	}
	for {
		repos, _, err := cl.c.ListMyRepos(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repos (page %d): %w", opts.Page, err)
		}
		all = append(all, repos...)
		if len(repos) < pageSize {
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
		ListOptions: gitea.ListOptions{PageSize: 50},
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
		ListOptions: gitea.ListOptions{PageSize: 50},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs for %s/%s: %w", owner, repoName, err)
	}
	return prs, nil
}

package gh

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/k1LoW/duration"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

const limit = 100
const defaultServerURL = "https://github.com"
const defaultGraphQLURL = "https://api.github.com/graphql"

type Client struct {
	v3            *github.Client
	v4            *githubv4.Client
	owner         string
	repo          string
	defaultBranch string
}

// New return Client
func New(ctx context.Context) (*Client, error) {
	// GITHUB_TOKEN
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("env %s is not set", "GITHUB_TOKEN")
	}

	// REST API Client
	v3c := github.NewClient(httpClient(token))
	if v3ep := os.Getenv("GITHUB_API_URL"); v3ep != "" {
		baseEndpoint, err := url.Parse(v3ep)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(baseEndpoint.Path, "/") {
			baseEndpoint.Path += "/"
		}
		v3c.BaseURL = baseEndpoint
	}

	// GraphQL API Client
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	v4hc := oauth2.NewClient(ctx, src)
	v4ep := os.Getenv("GITHUB_GRAPHQL_URL")
	if v4ep == "" {
		v4ep = defaultGraphQLURL
	}
	v4c := githubv4.NewEnterpriseClient(v4ep, v4hc)

	ownerrepo := os.Getenv("GITHUB_REPOSITORY")
	if ownerrepo == "" {
		return nil, fmt.Errorf("env %s is not set", "GITHUB_REPOSITORY")
	}
	splitted := strings.Split(ownerrepo, "/")

	owner := splitted[0]
	repo := splitted[1]

	r, _, err := v3c.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return &Client{
		v3:            v3c,
		v4:            v4c,
		owner:         owner,
		repo:          repo,
		defaultBranch: *r.DefaultBranch,
	}, nil
}

type PullRequestNode struct {
	Number      int
	Title       string
	URL         string
	Merged      bool
	MergedAt    time.Time
	UpdatedAt   time.Time
	MergeCommit struct {
		AbbreviatedOid string
	}
}

type PullRequestNodes []PullRequestNode

func (prs PullRequestNodes) Branch() string {
	numbers := []string{}
	for _, pr := range prs {
		numbers = append(numbers, strconv.Itoa(pr.Number))
	}
	return fmt.Sprintf("revert-%s-%d", strings.Join(numbers, "-"), time.Now().Unix())
}

func (prs PullRequestNodes) Title() string {
	numbers := []string{}
	for _, pr := range prs {
		numbers = append(numbers, fmt.Sprintf("#%d", pr.Number))
	}
	return fmt.Sprintf("Revert %s", strings.Join(numbers, " "))
}

func (prs PullRequestNodes) Body() string {
	numbers := []string{}
	for _, pr := range prs {
		if os.Getenv("GITHUB_SERVER_URL") == "" || os.Getenv("GITHUB_SERVER_URL") == defaultServerURL {
			// github.com
			numbers = append(numbers, fmt.Sprintf("- #%d", pr.Number))
		} else {
			numbers = append(numbers, fmt.Sprintf("- [**%s** #%d](%s)", pr.Title, pr.Number, pr.URL))
		}
	}
	footer := ""
	if os.Getenv("CI") != "" && os.Getenv("GITHUB_RUN_ID") != "" {
		footer = fmt.Sprintf("\n---\nCreated by %s/%s/actions/runs/%s\n", os.Getenv("GITHUB_SERVER_URL"), os.Getenv("GITHUB_REPOSITORY"), os.Getenv("GITHUB_RUN_ID"))
	}

	return fmt.Sprintf("Reverted pull requests:\n\n%s\n%s", strings.Join(numbers, "\n"), footer)
}

func (prs PullRequestNodes) Latest(l int) (PullRequestNodes, error) {
	if len(prs) < l {
		return nil, fmt.Errorf("there are not enough merged pull requests: %d < %d", len(prs), l)
	}
	return prs[:l], nil
}

func (prs PullRequestNodes) Before(b string) (PullRequestNodes, error) {
	d, err := duration.Parse(b)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	in := PullRequestNodes{}
	for _, pr := range prs {
		if now.Sub(pr.MergedAt).Nanoseconds() < d.Nanoseconds() {
			in = append(in, pr)
		}
	}
	if len(in) == 0 {
		return nil, fmt.Errorf("There were no PRs in the duration: %s", d.String())
	}
	return in, nil
}

func (c *Client) DefaultBranch() string {
	return c.defaultBranch
}

func (c *Client) FetchMergedPullRequest(ctx context.Context, n int) (PullRequestNode, error) {
	var q struct {
		Repogitory struct {
			PullRequest PullRequestNode `graphql:"pullRequest(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	variables := map[string]interface{}{
		"owner":  githubv4.String(c.owner),
		"repo":   githubv4.String(c.repo),
		"number": githubv4.Int(n),
	}
	if err := c.v4.Query(ctx, &q, variables); err != nil {
		return PullRequestNode{}, err
	}
	if !q.Repogitory.PullRequest.Merged {
		return PullRequestNode{}, fmt.Errorf("pull request #%d does not be merged", n)
	}
	return q.Repogitory.PullRequest, nil
}

func (c *Client) FetchMergedPullRequests(ctx context.Context) (PullRequestNodes, error) {
	var q struct {
		Repogitory struct {
			PullRequests struct {
				Nodes    PullRequestNodes
				PageInfo struct {
					HasNextPage bool
				}
			} `graphql:"pullRequests(first: $limit, states: [MERGED], orderBy: {direction: DESC, field: UPDATED_AT})"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	variables := map[string]interface{}{
		"owner": githubv4.String(c.owner),
		"repo":  githubv4.String(c.repo),
		"limit": githubv4.Int(limit),
	}

	if err := c.v4.Query(ctx, &q, variables); err != nil {
		return nil, err
	}

	return q.Repogitory.PullRequests.Nodes, nil
}

func (c *Client) CreatePullRequest(ctx context.Context, branch, title, body string) error {
	pr := &github.NewPullRequest{
		Title:               &title,
		Head:                &branch,
		Base:                &c.defaultBranch,
		Body:                &body,
		MaintainerCanModify: github.Bool(true),
	}
	if _, _, err := c.v3.PullRequests.Create(ctx, c.owner, c.repo, pr); err != nil {
		return err
	}

	return nil
}

type roundTripper struct {
	transport   *http.Transport
	accessToken string
}

func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", fmt.Sprintf("token %s", rt.accessToken))
	return rt.transport.RoundTrip(r)
}

func httpClient(token string) *http.Client {
	t := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	rt := roundTripper{
		transport:   t,
		accessToken: token,
	}
	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: rt,
	}
}

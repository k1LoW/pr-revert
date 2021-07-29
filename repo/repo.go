package repo

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

const defaultServerURL = "https://github.com"

type Repo struct {
	r *git.Repository
	d string
}

func New(ctx context.Context) (*Repo, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("env %s is not set", "GITHUB_TOKEN")
	}
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		return nil, fmt.Errorf("env %s is not set", "GITHUB_REPOSITORY")
	}
	host := os.Getenv("GITHUB_SERVER_URL")
	if host == "" {
		host = defaultServerURL
	}
	if _, err := exec.LookPath("git"); err != nil {
		return nil, err
	}

	d, err := os.MkdirTemp("", "pr-revert")
	if err != nil {
		return nil, err
	}
	u := strings.Replace(fmt.Sprintf("%s/%s.git", host, repo), "https://", fmt.Sprintf("https://%s:%s@", "dummy", token), 1)

	cmd := exec.CommandContext(ctx, "git", "clone", "--filter=tree:0", u, d)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	r, err := git.PlainOpen(d)
	if err != nil {
		return nil, err
	}

	return &Repo{
		r: r,
		d: d,
	}, nil
}

func (r *Repo) Dir() string {
	return r.d
}

func (r *Repo) Switch(ctx context.Context, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", r.d, "switch", "-c", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *Repo) RevertMergeCommit(ctx context.Context, oid string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", r.d, "revert", "-m", "1", oid)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *Repo) Push(ctx context.Context, branch string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("env %s is not set", "GITHUB_TOKEN")
	}
	cmd := exec.CommandContext(ctx, "git", "-C", r.d, "switch", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return r.r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "dummy",
			Password: token,
		},
		Progress: os.Stdout,
	})
}

func (r *Repo) Cleanup() error {
	log.Println("Cleanup repository")
	return os.RemoveAll(r.d)
}
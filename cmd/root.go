/*
Copyright © 2021 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/k1LoW/pr-revert/gh"
	"github.com/k1LoW/pr-revert/repo"
	"github.com/k1LoW/pr-revert/version"
	"github.com/spf13/cobra"
)

var (
	l         int
	u         string
	n         int
	noPush    bool
	noPR      bool
	noCleanup bool
	noBranch  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "pr-revert",
	Short:        "pr-revert is a tool for reverting pull requests",
	Long:         `pr-revert is a tool for reverting pull requests.`,
	Version:      version.Version,
	SilenceUsage: true,
	Args: func(cmd *cobra.Command, args []string) error {
		var err error
		if os.Getenv("PR_REVERT_LATEST") != "" && l == 0 {
			l, err = strconv.Atoi(os.Getenv("PR_REVERT_LATEST"))
			if err != nil {
				return err
			}
		}
		if os.Getenv("PR_REVERT_UNTIL") != "" && u == "" {
			u = os.Getenv("PR_REVERT_UNTIL")
		}
		if os.Getenv("PR_REVERT_NUMBER") != "" && n == 0 {
			n, err = strconv.Atoi(os.Getenv("PR_REVERT_NUMBER"))
			if err != nil {
				return err
			}
		}
		if os.Getenv("PR_REVERT_NO_PUSH") != "" && !noPush {
			noPush = true
		}
		if os.Getenv("PR_REVERT_NO_PULL_REQUEST") != "" && !noPR {
			noPR = true
		}
		if os.Getenv("PR_REVERT_NO_CLEANUP") != "" && !noCleanup {
			noCleanup = true
		}
		if os.Getenv("PR_REVERT_NO_BRANCH") != "" && !noBranch {
			noBranch = true
		}

		if l == 0 && u == "" && n == 0 {
			return errors.New("--latest (-l, PR_REVERT_LATEST) or --until (-u, PR_REVERT_UNTIL) or --number (-n, PR_REVERT_NUMBER) is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		r, err := repo.New(ctx)
		if err != nil {
			return err
		}
		if noCleanup {
			cmd.Printf("temporary working directory: %s\n", r.Dir())
		} else {
			defer r.Cleanup()
		}

		c, err := gh.New(ctx)
		if err != nil {
			return err
		}

		var prs gh.PullRequestNodes
		if n > 0 {
			pr, err := c.FetchMergedPullRequest(ctx, n)
			if err != nil {
				return err
			}
			prs = gh.PullRequestNodes{pr}
		} else {
			prs, err = c.FetchMergedPullRequests(ctx)
			if err != nil {
				return err
			}
		}

		if l > 0 {
			prs, err = prs.Latest(l)
			if err != nil {
				return err
			}
		}
		if u != "" {
			prs, err = prs.Before(u)
			if err != nil {
				return err
			}
		}

		var branch string

		if noBranch {
			branch = c.DefaultBranch()
		} else {
			branch = prs.Branch()
			if err := r.Switch(ctx, branch); err != nil {
				return err
			}
		}
		for _, pr := range prs {
			oid := pr.MergeCommit.AbbreviatedOid
			if err := r.RevertMergeCommit(ctx, oid); err != nil {
				return err
			}
		}
		if !noPush {
			if err := r.Push(ctx, branch); err != nil {
				return err
			}
			if !noPR {
				title := prs.Title()
				body := prs.Body()
				sig := prs.Sig()
				if err := c.CreatePullRequest(ctx, branch, title, body, sig); err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func Execute() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	log.SetOutput(io.Discard)
	if env := os.Getenv("DEBUG"); env != "" {
		debug, err := os.Create(fmt.Sprintf("%s.debug", version.Name))
		if err != nil {
			rootCmd.PrintErrln(err)
			os.Exit(1)
		}
		log.SetOutput(debug)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&l, "latest", "l", 0, "Number of most recently merged pull requests to revert")
	rootCmd.Flags().StringVarP(&u, "until", "u", "", "Duration of 'merged pull requests that will be reverted'")
	rootCmd.Flags().IntVarP(&n, "number", "n", 0, "Number of merged pull request to revert")
	rootCmd.Flags().BoolVarP(&noPush, "no-push", "", false, "Do not push branch")
	rootCmd.Flags().BoolVarP(&noPR, "no-pull-request", "", false, "Do not create a pull request")
	rootCmd.Flags().BoolVarP(&noCleanup, "no-cleanup", "", false, "Do not cleanup local repository")
	rootCmd.Flags().BoolVarP(&noBranch, "no-branch", "", false, "Do not create branch")
}

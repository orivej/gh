package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/google/go-github/github"
	"github.com/gregjones/httpcache"
	cli "github.com/jawher/mow.cli"
	"github.com/orivej/e"
	git "gopkg.in/src-d/go-git.v4"
	gitc "gopkg.in/src-d/go-git.v4/config"
	gitp "gopkg.in/src-d/go-git.v4/plumbing"
)

var rxGitHubURL = regexp.MustCompile(`(?i)github\.com[:/](.*?)/(.*?)(?:\.git)?$`)

func init() {
	app.Command("edit", "Edit GitHub pull request.", editCmd)
}

func editCmd(cmd *cli.Cmd) {
	pr := cmd.IntArg("PR", 0, "pull request number")
	prRemote := cmd.StringOpt("r remote", "origin", "remote of the pull request")
	action := func() int { return edit(*prRemote, *pr) }
	cmd.Action = func() { os.Exit(run(action)) }
}

func edit(prRemoteName string, prNumber int) int {
	// Open local git repo.
	cwd, err := os.Getwd()
	e.Exit(err)
	repo, err := git.PlainOpen(cwd)
	e.Exit(err)

	// Resolve PR repo.
	prRemote, err := repo.Remote(prRemoteName)
	e.Exit(err)
	url := prRemote.Config().URLs[0]
	parts := rxGitHubURL.FindStringSubmatch(url)
	prOwner, prRepo := parts[1], parts[2]

	// Fetch PR.
	log.Printf("* github: query %s/%s #%d", prOwner, prRepo, prNumber)
	ctx := context.Background()
	httpTransport := httpcache.NewMemoryCacheTransport()
	httpClient := http.Client{Transport: httpTransport}
	gh := github.NewClient(&httpClient)
	pr, _, err := gh.PullRequests.Get(ctx, prOwner, prRepo, prNumber)
	e.Exit(err)

	remoteName := *pr.Head.Repo.Owner.Login
	headRepo := *pr.Head.Repo.GitURL
	headBranch := *pr.Head.Ref
	localBranch := remoteName + "/" + headBranch
	localRefName := gitp.ReferenceName("refs/heads/" + localBranch)
	remoteHeadName := gitp.ReferenceName("refs/heads/" + headBranch)
	remoteRefName := gitp.ReferenceName("refs/remotes/" + localBranch)
	fetchRefSpec := gitc.RefSpec(fmt.Sprintf("+%s:%s", remoteHeadName, remoteRefName))

	// Add remote.
	_, err = repo.CreateRemote(&gitc.RemoteConfig{
		Name:  remoteName,
		URLs:  []string{headRepo},
		Fetch: []gitc.RefSpec{fetchRefSpec},
	})
	if err != git.ErrRemoteExists {
		e.Exit(err)
	}

	// Fetch PR.
	log.Printf("* git: fetch %s", localBranch)
	if pureGo {
		err = repo.Fetch(&git.FetchOptions{
			RemoteName: remoteName,
			RefSpecs:   []gitc.RefSpec{fetchRefSpec},
		})
		if err != git.NoErrAlreadyUpToDate {
			e.Exit(err)
		}
	} else {
		runGit("fetch", "-u", remoteName, string(fetchRefSpec))
	}

	log.Printf("* git: checkout %s", localBranch)

	// Configure branch.
	runGit("branch", "-f", localBranch, string(remoteRefName))

	// Check it out.
	if pureGo {
		wt, err := repo.Worktree()
		e.Exit(err)
		err = wt.Checkout(&git.CheckoutOptions{Branch: localRefName})
		if err == git.ErrWorktreeNotClean {
			log.Print("checkout failed: working tree is not clean")
			return 1
		}
		e.Exit(err)
	} else {
		runGit("checkout", localBranch)
	}
	return 0
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	git "github.com/go-git/go-git"
	gitc "github.com/go-git/go-git/config"
	gitp "github.com/go-git/go-git/plumbing"
	"github.com/google/go-github/github"
	"github.com/gregjones/httpcache"
	cli "github.com/jawher/mow.cli"
	"github.com/orivej/e"
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
	if err == git.ErrRepositoryNotExists {
		log.Println("error: this is not the root of a git repository:", cwd)
		return 1
	}
	e.Exit(err)

	// Resolve PR repo.
	prRemote, err := repo.Remote(prRemoteName)
	e.Exit(err)
	url := prRemote.Config().URLs[0]
	replacer, err := NewGitURLReplacer()
	e.Exit(err)
	url, unreplace := replacer.Replace(url)
	parts := rxGitHubURL.FindStringSubmatch(url)
	if len(parts) != 3 {
		log.Printf("error: remote %q is not at GitHub: %s", prRemoteName, url)
		return 1
	}
	prOwner, prRepo := parts[1], parts[2]

	// Fetch PR.
	log.Printf("* github: query %s/%s #%d", prOwner, prRepo, prNumber)
	ctx := context.Background()
	httpTransport := httpcache.NewMemoryCacheTransport()
	httpClient := http.Client{Transport: httpTransport}
	gh := github.NewClient(&httpClient)
	pr, _, err := gh.PullRequests.Get(ctx, prOwner, prRepo, prNumber)
	e.Exit(err)

	headRepo := *pr.Head.Repo.GitURL
	remoteName := *pr.Head.Repo.Owner.Login
	remoteBranch := *pr.Head.Ref
	// localBranch (in refs/heads/) and shadowBranch (in refs/remotes/) are
	// configurable, everything else is determined.  For convenience of git
	// operations they should be different.  For pushRemote to work, the
	// localBranch must equal remoteBranch.
	shadowBranch := remoteName + "/" + remoteBranch
	localBranch := remoteBranch
	if localBranch == *pr.Base.Ref || localBranch == *pr.Base.Repo.DefaultBranch {
		localBranch = fmt.Sprintf("pr-%d", prNumber)
	}
	remoteRefName := gitp.ReferenceName("refs/heads/" + remoteBranch)
	shadowRefName := gitp.ReferenceName("refs/remotes/" + shadowBranch)
	localRefName := gitp.ReferenceName("refs/heads/" + localBranch)
	fetchRefSpec := gitc.RefSpec(fmt.Sprintf("+%s:%s", remoteRefName, shadowRefName))

	// Add remote.
	config := &gitc.RemoteConfig{
		Name:  remoteName,
		URLs:  []string{unreplace(headRepo)},
		Fetch: []gitc.RefSpec{fetchRefSpec},
	}
	remote, err := repo.Remote(remoteName)
	if err == nil {
		config = remote.Config()
		for _, refspec := range config.Fetch {
			if refspec == fetchRefSpec {
				config = nil
				break
			}
		}
		if config != nil {
			config.Fetch = append(config.Fetch, fetchRefSpec)
			err = repo.DeleteRemote(remoteName)
			e.Exit(err)
		}
	}
	if config != nil {
		_, err = repo.CreateRemote(config)
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
	runGit("branch", "-f", localBranch, string(shadowRefName))

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

package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
)

func init() {
	app.Command("sync", "Update upstream and rebase onto its branch.", func(cmd *cli.Cmd) {
		remote := cmd.StringOpt("r remote", "origin", "Upstream remote.")
		branch := cmd.StringOpt("b branch", "master", "Upstream branch.")
		action := func() int { return sync(*remote, *branch) }
		cmd.Action = func() { os.Exit(run(action)) }
	})
}

func sync(remote, branch string) int {
	runGit("fetch", remote)
	// runGit("branch", "-f", "master", "refs/remotes/origin/master")
	// runGit("branch", "-f", "staging", "refs/remotes/origin/staging")
	// runGit("branch", "-f", "release-17.09", "refs/remotes/origin/release-17.09")
	ref := fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
	runGit("branch", "-f", branch, ref)
	runGit("rebase", ref)
	return 0
}

# Introduction

`gh` facilitates repository maintainers in working with pull requests.

# Installation

With `go`: `go get github.com/orivej/gh`

With `nix`: `nix-env -if https://github.com/orivej/gh/archive/master.tar.gz`

# Usage

`gh edit 1234` checks out PR #1234. If the PR author allows edits, you can edit it and publish your changes with `git push`.

`gh sync` fetches the upstream branch (which defaults to `master`) and rebases the current branch on top of the updated upstream.

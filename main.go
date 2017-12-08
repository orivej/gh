package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jawher/mow.cli"
	"github.com/orivej/e"
)

var (
	app = cli.App("gh", "GitHub automation.")

	binGit = "git"
	pureGo = false // "git checkout" is faster than go-git checkout.

	abort0 = 0
	abort  = &abort0
)

func main() {
	log.SetFlags(0)
	err := app.Run(os.Args)
	e.Exit(err)
}

func runGit(args ...string) {
	cmd := exec.Command(binGit, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	log.Println("+", binGit, strings.Join(args, " "))
	err := cmd.Run()
	if err != nil {
		panic(abort)
	}
}

func run(f func() int) (rc int) {
	defer func() {
		v := recover()
		if v == abort {
			rc = 1
		} else if v != nil {
			panic(v)
		}
	}()
	rc = f()
	return
}

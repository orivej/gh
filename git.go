package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type GitURLReplacer map[string]string

func NewGitURLReplacer() (GitURLReplacer, error) {
	cmd := exec.Command(binGit, "config", "--get-regexp", `^url[.].*[.]insteadOf$`)
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to invoke git config")
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	g := GitURLReplacer{}
	for scanner.Scan() {
		var url, prefix string
		_, err = fmt.Sscanf(scanner.Text(), "url.%s %s", &url, &prefix)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse git config")
		}
		g[prefix] = strings.TrimSuffix(url, ".insteadof")
	}
	return g, nil
}

func (g GitURLReplacer) Replace(s string) (string, func(string) string) {
	for prefix, url := range g {
		if strings.HasPrefix(s, prefix) {
			return url + s[len(prefix):], func(s string) string {
				if strings.HasPrefix(s, url) {
					return prefix + s[len(url):]
				}
				return s
			}
		}
	}
	return s, func(s string) string { return s }
}

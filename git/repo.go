package git

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type repo struct {
	url       string
	deployKey string
	dir       string
}

func NewRepo(url, deployKey string, dir string) *repo {
	return &repo{
		url:       url,
		deployKey: deployKey,
		dir:       dir,
	}
}

func (r *repo) CloneOrPull() error {
	var auth transport.AuthMethod
	if r.deployKey != "" {
		if block, _ := pem.Decode([]byte(r.deployKey)); block == nil {
			return errors.New("failed to parse public key: invalid PEM")
		}
		var err error
		auth, err = ssh.NewPublicKeys("git", []byte(r.deployKey), "")
		if err != nil {
			return errors.Wrap(err, "failed to parse public key")
		}
	}

	if r.cloned() {
		repo, err := git.PlainOpen(r.dir)
		if err != nil {
			return errors.Wrap(err, "failed to open repo")
		}
		if err := repo.Fetch(&git.FetchOptions{Auth: auth}); err != nil && err != git.NoErrAlreadyUpToDate {
			return errors.Wrap(err, "failed to fetch repo")
		}
		if output, err := r.run("reset", "--hard", "origin/master"); err != nil {
			return errors.Errorf("failed to reset repo: %s: %s", err, string(output))
		}
	} else {
		_, err := git.PlainClone(r.dir, false, &git.CloneOptions{URL: r.url, Auth: auth})
		if err != nil {
			return errors.Wrap(err, "failed to clone repo")
		}
	}

	return nil
}

func (r *repo) CommitAndPush(message, committer string) error {
	if output, err := r.run("add", "-A"); err != nil {
		return errors.Errorf("failed to stage files: %s: %s", err, string(output))
	}
	if output, err := r.run("-c", "user.name=Claimer", "-c", "user.email=<>", "commit", "--author", committer+" <>", "-m", message); err != nil {
		return errors.Errorf("failed to commit: %s: %s", err, string(output))
	}
	if output, err := r.run("push", "origin", "master"); err != nil {
		return errors.Errorf("failed to push: %s: %s", err, string(output))
	}
	return nil
}

func (r *repo) Dir() string {
	return r.dir
}

func (r *repo) LatestCommit(path string) (string, string, string, error) {
	author, err := r.run("log", "-1", "--format=%an", path)
	if err != nil {
		return "", "", "", errors.Errorf("failed to get commit author: %s: %s", err, string(author))
	}
	date, err := r.run("log", "-1", "--format=%ad", path)
	if err != nil {
		return "", "", "", errors.Errorf("failed to get commit date: %s: %s", err, string(date))
	}
	body, err := r.run("log", "-1", "--format=%b", path)
	if err != nil {
		return "", "", "", errors.Errorf("failed to get commit body: %s: %s", err, string(date))
	}

	return strings.TrimSpace(string(author)), strings.TrimSpace(string(date)), strings.TrimSpace(string(body)), nil
}

func (r *repo) cloned() bool {
	output, err := r.run("rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(string(output)) == "true"
}

func (r *repo) run(args ...string) ([]byte, error) {
	tempDir, err := ioutil.TempDir("", "claimer-git")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	deployKeyPath := filepath.Join(tempDir, "deployKey")
	if err := ioutil.WriteFile(deployKeyPath, []byte(r.deployKey), 0600); err != nil {
		return nil, err
	}

	sshPath := filepath.Join(tempDir, "ssh")
	sshScript := fmt.Sprintf("#!/bin/sh\n"+
		`exec /usr/bin/ssh -o StrictHostKeyChecking=no -i %s "$@"`, deployKeyPath)
	if err := ioutil.WriteFile(sshPath, []byte(sshScript), 0755); err != nil {
		return nil, err
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH=%s", sshPath))
	return cmd.CombinedOutput()
}

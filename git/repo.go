package git

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"os/exec"
	"srcd.works/go-git.v4"
	"srcd.works/go-git.v4/plumbing/transport"
	gitssh "srcd.works/go-git.v4/plumbing/transport/ssh"
	"strings"
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
		signer, err := ssh.ParsePrivateKey([]byte(r.deployKey))
		if err != nil {
			return fmt.Errorf("failed to parse public key: %s", err)
		}

		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}

	if r.cloned() {
		r, err := git.PlainOpen(r.dir)
		if err != nil {
			return fmt.Errorf("failed to open repo: %s", err)
		}
		if err := r.Pull(&git.PullOptions{Auth: auth}); err != nil {
			return fmt.Errorf("failed to pull repo: %s", err)
		}
	} else {
		_, err := git.PlainClone(r.dir, false, &git.CloneOptions{URL: r.url, Auth: auth})
		if err != nil {
			return fmt.Errorf("failed to clone repo: %s", err)
		}
	}

	return nil
}

func (r *repo) cloned() bool {
	output, err := r.run("rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(string(output)) == "true"
}

func (r *repo) CommitAndPush(message string) error {
	if output, err := r.run("add", "-A"); err != nil {
		return fmt.Errorf("failed to stage files: %s", string(output))
	}
	if output, err := r.run("commit", "-m", message); err != nil {
		return fmt.Errorf("failed to commit: %s", string(output))
	}
	if output, err := r.run("push", "origin", "master"); err != nil {
		return fmt.Errorf("failed to push: %s", string(output))
	}
	return nil
}

func (r *repo) run(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	return cmd.CombinedOutput()
}

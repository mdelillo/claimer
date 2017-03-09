package git

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"os/exec"
	"srcd.works/go-git.v4"
	"srcd.works/go-git.v4/plumbing/transport"
	gitssh "srcd.works/go-git.v4/plumbing/transport/ssh"
)

type repo struct {
	dir string
}

func NewRepo(url, deployKey string) (*repo, error) {
	var auth transport.AuthMethod
	if deployKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(deployKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %s", err)
		}

		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}

	dir, err := ioutil.TempDir("", "claimer-git-repo")
	if err != nil {
		return nil, err
	}

	_, err = git.PlainClone(dir, false, &git.CloneOptions{URL: url, Auth: auth})
	if err != nil {
		os.RemoveAll(dir)
		return nil, fmt.Errorf("failed to clone repo: %s", err)
	}

	return &repo{dir: dir}, nil
}

func (r *repo) Dir() string {
	return r.dir
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
	cmd.Dir = r.Dir()
	return cmd.CombinedOutput()
}

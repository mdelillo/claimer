package git

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os/exec"
	"srcd.works/go-git.v4"
	gitssh "srcd.works/go-git.v4/plumbing/transport/ssh"
)

type repo struct {
	Dir string
}

func NewRepo(url, deployKey string) *repo {
	dir, err := ioutil.TempDir("", "claimer-git-repo")
	if err != nil {
		panic(err)
	}

	var auth *gitssh.PublicKeys
	if deployKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(deployKey))
		if err != nil {
			panic(err)
		}

		auth.User = "git"
		auth.Signer = signer
	}

	_, err = git.PlainClone(dir, false, &git.CloneOptions{URL: url, Auth: auth})
	if err != nil {
		panic(err)
	}

	return &repo{
		Dir: dir,
	}
}

func (r *repo) CommitAndPush(message string) {
	if output, err := r.run("add", "-A"); err != nil {
		panic(string(output))
	}
	if output, err := r.run("commit", "-m", message); err != nil {
		panic(string(output))
	}
	if output, err := r.run("push", "origin", "master"); err != nil {
		panic(string(output))
	}
}

func (r *repo) run(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Dir
	return cmd.CombinedOutput()
}

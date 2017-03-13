package locker

import (
	"fmt"
	"path/filepath"
)

//go:generate counterfeiter . gitRepo
type gitRepo interface {
	CloneOrPull() error
	CommitAndPush(message string) error
	Dir() string
}

//go:generate counterfeiter . fs
type fs interface {
	Ls(dir string) ([]string, error)
	Mv(src, dst string) error
}

type locker struct {
	fs      fs
	gitRepo gitRepo
}

func NewLocker(fs fs, gitRepo gitRepo) *locker {
	return &locker{
		fs:      fs,
		gitRepo: gitRepo,
	}
}

func (l *locker) ClaimLock(pool string) error {
	if err := l.gitRepo.CloneOrPull(); err != nil {
		return err
	}

	locks, err := l.fs.Ls(filepath.Join(l.gitRepo.Dir(), pool, "unclaimed"))
	if err != nil {
		return err
	}

	if len(locks) == 0 {
		return fmt.Errorf("no unclaimed locks for pool " + pool)
	} else if len(locks) > 1 {
		return fmt.Errorf("too many unclaimed locks for pool " + pool)
	}

	unclaimedLock := filepath.Join(l.gitRepo.Dir(), pool, "unclaimed", locks[0])
	claimedLock := filepath.Join(l.gitRepo.Dir(), pool, "claimed", locks[0])
	if err := l.fs.Mv(unclaimedLock, claimedLock); err != nil {
		return err
	}

	if err := l.gitRepo.CommitAndPush("Claimer claiming " + pool); err != nil {
		return err
	}
	return nil
}

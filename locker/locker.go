package locker

import (
	"fmt"
	"path/filepath"
)

//go:generate counterfeiter . gitRepo
type gitRepo interface {
	CloneOrPull() error
	CommitAndPush(message, user string) error
	Dir() string
	LatestCommit(pool string) (committer, date string, err error)
}

//go:generate counterfeiter . fs
type fs interface {
	Ls(dir string) ([]string, error)
	LsDirs(dir string) ([]string, error)
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

func (l *locker) ClaimLock(pool, user string) error {
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

	if err := l.gitRepo.CommitAndPush("Claimer claiming " + pool, user); err != nil {
		return err
	}
	return nil
}

func (l *locker) Owner(pool string) (string, string, error) {
	if err := l.gitRepo.CloneOrPull(); err != nil {
		return "", "", err
	}

	return l.gitRepo.LatestCommit(pool)
}

func (l *locker) ReleaseLock(pool, user string) error {
	if err := l.gitRepo.CloneOrPull(); err != nil {
		return err
	}

	locks, err := l.fs.Ls(filepath.Join(l.gitRepo.Dir(), pool, "claimed"))
	if err != nil {
		return err
	}

	if len(locks) == 0 {
		return fmt.Errorf("no claimed locks for pool " + pool)
	} else if len(locks) > 1 {
		return fmt.Errorf("too many claimed locks for pool " + pool)
	}

	claimedLock := filepath.Join(l.gitRepo.Dir(), pool, "claimed", locks[0])
	unclaimedLock := filepath.Join(l.gitRepo.Dir(), pool, "unclaimed", locks[0])
	if err := l.fs.Mv(claimedLock, unclaimedLock); err != nil {
		return err
	}

	if err := l.gitRepo.CommitAndPush("Claimer releasing " + pool, user); err != nil {
		return err
	}
	return nil
}

func (l *locker) Status() ([]string, []string, error) {
	var claimedPools []string
	var unclaimedPools []string

	if err := l.gitRepo.CloneOrPull(); err != nil {
		return nil, nil, err
	}

	pools, err := l.fs.LsDirs(l.gitRepo.Dir())
	if err != nil {
		return nil, nil, err
	}
	for _, pool := range pools {
		claimedLocks, err := l.fs.Ls(filepath.Join(l.gitRepo.Dir(), pool, "claimed"))
		if err != nil {
			return nil, nil, err
		}
		if len(claimedLocks) > 0 {
			claimedPools = append(claimedPools, pool)
		} else {
			unclaimedLocks, err := l.fs.Ls(filepath.Join(l.gitRepo.Dir(), pool, "unclaimed"))
			if err != nil {
				return nil, nil, err
			}
			if len(unclaimedLocks) > 0 {
				unclaimedPools = append(unclaimedPools, pool)
			}
		}
	}

	return claimedPools, unclaimedPools, nil
}

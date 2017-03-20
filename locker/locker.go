package locker

import (
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "failed to clone or pull")
	}

	locks, err := l.fs.Ls(filepath.Join(l.gitRepo.Dir(), pool, "unclaimed"))
	if err != nil {
		return errors.Wrap(err, "failed to list unclaimed locks")
	}

	if len(locks) == 0 {
		return errors.Errorf("no unclaimed locks for pool %s", pool)
	} else if len(locks) > 1 {
		return errors.Errorf("too many unclaimed locks for pool %s", pool)
	}

	unclaimedLock := filepath.Join(l.gitRepo.Dir(), pool, "unclaimed", locks[0])
	claimedLock := filepath.Join(l.gitRepo.Dir(), pool, "claimed", locks[0])
	if err := l.fs.Mv(unclaimedLock, claimedLock); err != nil {
		return errors.Wrap(err, "failed to move file")
	}

	if err := l.gitRepo.CommitAndPush("Claimer claiming "+pool, user); err != nil {
		return errors.Wrap(err, "failed to commit and push")
	}
	return nil
}

func (l *locker) Owner(pool string) (string, string, error) {
	if err := l.gitRepo.CloneOrPull(); err != nil {
		return "", "", errors.Wrap(err, "failed to clone or pull")
	}

	author, date, err := l.gitRepo.LatestCommit(pool)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get latest commit")
	}
	return author, date, nil
}

func (l *locker) ReleaseLock(pool, user string) error {
	if err := l.gitRepo.CloneOrPull(); err != nil {
		return errors.Wrap(err, "failed to clone or pull")
	}

	locks, err := l.fs.Ls(filepath.Join(l.gitRepo.Dir(), pool, "claimed"))
	if err != nil {
		return errors.Wrap(err, "failed to list claimed locks")
	}

	if len(locks) == 0 {
		return errors.Errorf("no claimed locks for pool %s", pool)
	} else if len(locks) > 1 {
		return errors.Errorf("too many claimed locks for pool %s", pool)
	}

	claimedLock := filepath.Join(l.gitRepo.Dir(), pool, "claimed", locks[0])
	unclaimedLock := filepath.Join(l.gitRepo.Dir(), pool, "unclaimed", locks[0])
	if err := l.fs.Mv(claimedLock, unclaimedLock); err != nil {
		return errors.Wrap(err, "failed to move file")
	}

	if err := l.gitRepo.CommitAndPush("Claimer releasing "+pool, user); err != nil {
		return errors.Wrap(err, "failed to commit and push")
	}
	return nil
}

func (l *locker) Status() ([]string, []string, error) {
	var claimedPools []string
	var unclaimedPools []string

	if err := l.gitRepo.CloneOrPull(); err != nil {
		return nil, nil, errors.Wrap(err, "failed to clone or pull")
	}

	pools, err := l.fs.LsDirs(l.gitRepo.Dir())
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list pools")
	}
	for _, pool := range pools {
		claimedLocks, err := l.fs.Ls(filepath.Join(l.gitRepo.Dir(), pool, "claimed"))
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to list claimed locks")
		}
		if len(claimedLocks) > 0 {
			claimedPools = append(claimedPools, pool)
		} else {
			unclaimedLocks, err := l.fs.Ls(filepath.Join(l.gitRepo.Dir(), pool, "unclaimed"))
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to list unclaimed locks")
			}
			if len(unclaimedLocks) > 0 {
				unclaimedPools = append(unclaimedPools, pool)
			}
		}
	}

	return claimedPools, unclaimedPools, nil
}

package locker_test

import (
	. "github.com/mdelillo/claimer/locker"

	"errors"
	"github.com/mdelillo/claimer/locker/lockerfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"path/filepath"
)

var _ = Describe("Locker", func() {
	var (
		fs      *lockerfakes.FakeFs
		gitRepo *lockerfakes.FakeGitRepo
	)

	BeforeEach(func() {
		fs = new(lockerfakes.FakeFs)
		gitRepo = new(lockerfakes.FakeGitRepo)
	})

	Describe("ClaimLock", func() {
		It("claims the lock file in the git repo", func() {
			pool := "some-pool"
			gitDir := "some-dir"
			lock := "some-lock"
			user := "some-user"

			gitRepo.DirReturns(gitDir)
			fs.LsReturns([]string{lock}, nil)

			locker := NewLocker(fs, gitRepo)
			Expect(locker.ClaimLock(pool, user)).To(Succeed())

			Expect(gitRepo.CloneOrPullCallCount()).To(Equal(1))

			Expect(fs.LsCallCount()).To(Equal(1))
			Expect(fs.LsArgsForCall(0)).To(Equal(filepath.Join(gitDir, pool, "unclaimed")))

			Expect(fs.MvCallCount()).To(Equal(1))
			oldPath, newPath := fs.MvArgsForCall(0)
			Expect(oldPath).To(Equal(filepath.Join(gitDir, pool, "unclaimed", lock)))
			Expect(newPath).To(Equal(filepath.Join(gitDir, pool, "claimed", lock)))

			message, actualUser := gitRepo.CommitAndPushArgsForCall(0)
			Expect(gitRepo.CommitAndPushCallCount()).To(Equal(1))
			Expect(message).To(Equal("Claimer claiming " + pool))
			Expect(actualUser).To(Equal(user))
		})

		Context("when cloning the repo fails", func() {
			It("returns an error", func() {
				gitRepo.CloneOrPullReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock("", "")).To(MatchError("some-error"))
			})
		})

		Context("when listing files fails", func() {
			It("returns an error", func() {
				fs.LsReturns(nil, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock("", "")).To(MatchError("some-error"))
			})
		})

		Context("when there are no unclaimed locks", func() {
			It("returns an error", func() {
				pool := "some-pool"

				fs.LsReturns([]string{}, nil)

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock(pool, "")).To(MatchError("no unclaimed locks for pool " + pool))
			})
		})

		Context("when there are multiple unclaimed locks", func() {
			It("returns an error", func() {
				pool := "some-pool"

				fs.LsReturns([]string{"some-lock", "some-other-lock"}, nil)

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock(pool, "")).To(MatchError("too many unclaimed locks for pool " + pool))
			})
		})

		Context("when moving the file fails", func() {
			It("returns an error", func() {
				fs.LsReturns([]string{"some-lock"}, nil)
				fs.MvReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock("", "")).To(MatchError("some-error"))
			})
		})

		Context("when pushing fails", func() {
			It("returns an error", func() {
				fs.LsReturns([]string{"some-lock"}, nil)
				gitRepo.CommitAndPushReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock("", "")).To(MatchError("some-error"))
			})
		})
	})

	Describe("Owner", func() {
		It("returns the author and date of the latest commit to the pool", func() {
			pool := "some-pool"
			author := "some-author"
			date := "some-date"

			gitRepo.LatestCommitReturns(author, date, nil)

			locker := NewLocker(fs, gitRepo)
			owner, actualDate, err := locker.Owner(pool)
			Expect(err).NotTo(HaveOccurred())

			Expect(gitRepo.CloneOrPullCallCount()).To(Equal(1))

			Expect(owner).To(Equal(author))
			Expect(actualDate).To(Equal(date))
		})

		Context("when cloning the repo fails", func() {
			It("returns an error", func() {
				gitRepo.CloneOrPullReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				_, _, err := locker.Owner("")
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when getting the author fails", func() {
			It("returns an error", func() {
				gitRepo.LatestCommitReturns("", "", errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				_, _, err := locker.Owner("")
				Expect(err).To(MatchError("some-error"))
			})
		})
	})

	Describe("ReleaseLock", func() {
		It("releases the lock file in the git repo", func() {
			pool := "some-pool"
			gitDir := "some-dir"
			lock := "some-lock"
			user := "some-user"

			gitRepo.DirReturns(gitDir)
			fs.LsReturns([]string{lock}, nil)

			locker := NewLocker(fs, gitRepo)
			Expect(locker.ReleaseLock(pool, user)).To(Succeed())

			Expect(gitRepo.CloneOrPullCallCount()).To(Equal(1))

			Expect(fs.LsCallCount()).To(Equal(1))
			Expect(fs.LsArgsForCall(0)).To(Equal(filepath.Join(gitDir, pool, "claimed")))

			Expect(fs.MvCallCount()).To(Equal(1))
			oldPath, newPath := fs.MvArgsForCall(0)
			Expect(oldPath).To(Equal(filepath.Join(gitDir, pool, "claimed", lock)))
			Expect(newPath).To(Equal(filepath.Join(gitDir, pool, "unclaimed", lock)))

			message, actualUser := gitRepo.CommitAndPushArgsForCall(0)
			Expect(gitRepo.CommitAndPushCallCount()).To(Equal(1))
			Expect(message).To(Equal("Claimer releasing " + pool))
			Expect(actualUser).To(Equal(user))
		})

		Context("when cloning the repo fails", func() {
			It("returns an error", func() {
				gitRepo.CloneOrPullReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ReleaseLock("", "")).To(MatchError("some-error"))
			})
		})

		Context("when listing files fails", func() {
			It("returns an error", func() {
				fs.LsReturns(nil, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ReleaseLock("", "")).To(MatchError("some-error"))
			})
		})

		Context("when there are no claimed locks", func() {
			It("returns an error", func() {
				pool := "some-pool"

				fs.LsReturns([]string{}, nil)

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ReleaseLock(pool, "")).To(MatchError("no claimed locks for pool " + pool))
			})
		})

		Context("when there are multiple claimed locks", func() {
			It("returns an error", func() {
				pool := "some-pool"

				fs.LsReturns([]string{"some-lock", "some-other-lock"}, nil)

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ReleaseLock(pool, "")).To(MatchError("too many claimed locks for pool " + pool))
			})
		})

		Context("when moving the file fails", func() {
			It("returns an error", func() {
				fs.LsReturns([]string{"some-lock"}, nil)
				fs.MvReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ReleaseLock("", "")).To(MatchError("some-error"))
			})
		})

		Context("when pushing fails", func() {
			It("returns an error", func() {
				fs.LsReturns([]string{"some-lock"}, nil)
				gitRepo.CommitAndPushReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ReleaseLock("", "")).To(MatchError("some-error"))
			})
		})
	})

	Describe("Status", func() {
		It("returns lists of claimed and unclaimed pools", func() {
			gitDir := "some-dir"
			gitRepo.DirReturns(gitDir)

			fs.LsDirsStub = func(dir string) ([]string, error) {
				if dir == gitDir {
					return []string{"pool-1", "pool-2", "empty-pool", "ful-pool"}, nil
				} else {
					return []string{}, nil
				}
			}
			fs.LsStub = func(dir string) ([]string, error) {
				if dir == filepath.Join(gitDir, "pool-1", "claimed") {
					return []string{"lock"}, nil
				} else if dir == filepath.Join(gitDir, "pool-2", "unclaimed") {
					return []string{"lock"}, nil
				} else if dir == filepath.Join(gitDir, "full-pool", "claimed") {
					return []string{"lock"}, nil
				} else if dir == filepath.Join(gitDir, "full-pool", "unclaimed") {
					return []string{"lock"}, nil
				} else {
					return []string{}, nil
				}
			}

			locker := NewLocker(fs, gitRepo)
			claimedPools, unclaimedPools, err := locker.Status()
			Expect(err).NotTo(HaveOccurred())
			Expect(claimedPools).To(Equal([]string{"pool-1"}))
			Expect(unclaimedPools).To(Equal([]string{"pool-2"}))

			Expect(gitRepo.CloneOrPullCallCount()).To(Equal(1))
		})

		Context("when cloning the repo fails", func() {
			It("returns an error", func() {
				gitRepo.CloneOrPullReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				_, _, err := locker.Status()
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when listing the git repo fails", func() {
			It("returns an error", func() {
				fs.LsDirsReturns(nil, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				_, _, err := locker.Status()
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when listing claimed locks fails", func() {
			It("returns an error", func() {
				fs.LsDirsReturns([]string{"some-pool"}, nil)
				fs.LsReturns(nil, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				_, _, err := locker.Status()
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when listing unclaimed locks fails", func() {
			It("returns an error", func() {
				fs.LsDirsReturns([]string{"some-pool"}, nil)
				fs.LsReturnsOnCall(1, nil, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				_, _, err := locker.Status()
				Expect(err).To(MatchError("some-error"))
			})
		})
	})
})

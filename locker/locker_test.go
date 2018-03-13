package locker_test

import (
	. "github.com/mdelillo/claimer/locker"

	"errors"
	"fmt"
	"path/filepath"

	"github.com/mdelillo/claimer/locker/lockerfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		Context("when the lock is just a pool", func() {
			It("claims the lock file in the git repo", func() {
				pool := "some-pool"
				gitDir := "some-dir"
				lock := "some-lock"
				user := "some-user"
				message := "some-message"

				gitRepo.DirReturns(gitDir)
				fs.LsReturns([]string{lock}, nil)

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock(pool, user, message)).To(Succeed())

				Expect(gitRepo.CloneOrPullCallCount()).To(Equal(1))

				Expect(fs.LsCallCount()).To(Equal(1))
				Expect(fs.LsArgsForCall(0)).To(Equal(filepath.Join(gitDir, pool, "unclaimed")))

				Expect(fs.MvCallCount()).To(Equal(1))
				oldPath, newPath := fs.MvArgsForCall(0)
				Expect(oldPath).To(Equal(filepath.Join(gitDir, pool, "unclaimed", lock)))
				Expect(newPath).To(Equal(filepath.Join(gitDir, pool, "claimed", lock)))

				actualMessage, actualUser := gitRepo.CommitAndPushArgsForCall(0)
				Expect(gitRepo.CommitAndPushCallCount()).To(Equal(1))
				Expect(actualMessage).To(Equal(fmt.Sprintf("Claimer claiming %s\n\n%s", pool, message)))
				Expect(actualUser).To(Equal(user))
			})
		})

		FContext("when the lock is a pool/lock", func() {
			It("claims the lock file in the git repo", func() {
				pool := "some-pool"
				lock := "some-lock-1"
				otherLock := "some-lock-2"
				gitDir := "some-dir"
				user := "some-user"
				message := "some-message"

				gitRepo.DirReturns(gitDir)
				fs.LsReturns([]string{lock, otherLock}, nil)

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock(pool+"/"+lock, user, message)).To(Succeed())

				Expect(gitRepo.CloneOrPullCallCount()).To(Equal(1))

				Expect(fs.LsCallCount()).To(Equal(1))
				Expect(fs.LsArgsForCall(0)).To(Equal(filepath.Join(gitDir, pool, "unclaimed")))

				Expect(fs.MvCallCount()).To(Equal(1))
				oldPath, newPath := fs.MvArgsForCall(0)
				Expect(oldPath).To(Equal(filepath.Join(gitDir, pool, "unclaimed", lock)))
				Expect(newPath).To(Equal(filepath.Join(gitDir, pool, "claimed", lock)))

				actualMessage, actualUser := gitRepo.CommitAndPushArgsForCall(0)
				Expect(gitRepo.CommitAndPushCallCount()).To(Equal(1))
				Expect(actualMessage).To(Equal(fmt.Sprintf("Claimer claiming %s/%s\n\n%s", pool, lock, message)))
				Expect(actualUser).To(Equal(user))
			})
		})

		Context("when the message is empty", func() {
			It("claims the lock file without an extra message", func() {
				pool := "some-pool"
				gitDir := "some-dir"
				lock := "some-lock"
				user := "some-user"

				gitRepo.DirReturns(gitDir)
				fs.LsReturns([]string{lock}, nil)

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock(pool, user, "")).To(Succeed())

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
		})

		Context("when cloning the repo fails", func() {
			It("returns an error", func() {
				gitRepo.CloneOrPullReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock("", "", "")).To(MatchError("failed to clone or pull: some-error"))
			})
		})

		Context("when listing files fails", func() {
			It("returns an error", func() {
				fs.LsReturns(nil, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock("", "", "")).To(MatchError("failed to list unclaimed locks: some-error"))
			})
		})

		Context("when there are no unclaimed locks", func() {
			It("returns an error", func() {
				pool := "some-pool"

				fs.LsReturns([]string{}, nil)

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock(pool, "", "")).To(MatchError("no unclaimed locks for pool " + pool))
			})
		})

		Context("when there are multiple unclaimed locks", func() {
			It("returns an error", func() {
				pool := "some-pool"

				fs.LsReturns([]string{"some-lock", "some-other-lock"}, nil)

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock(pool, "", "")).To(MatchError("too many unclaimed locks for pool " + pool))
			})
		})

		Context("when moving the file fails", func() {
			It("returns an error", func() {
				fs.LsReturns([]string{"some-lock"}, nil)
				fs.MvReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock("", "", "")).To(MatchError("failed to move file: some-error"))
			})
		})

		Context("when pushing fails", func() {
			It("returns an error", func() {
				fs.LsReturns([]string{"some-lock"}, nil)
				gitRepo.CommitAndPushReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ClaimLock("", "", "")).To(MatchError("failed to commit and push: some-error"))
			})
		})
	})

	Describe("CreatePool", func() {
		It("creates a pool with an unclaimed lock", func() {
			pool := "some-pool"
			gitDir := "some-dir"
			user := "some-user"

			gitRepo.DirReturns(gitDir)

			locker := NewLocker(fs, gitRepo)
			Expect(locker.CreatePool(pool, user)).To(Succeed())

			Expect(gitRepo.CloneOrPullCallCount()).To(Equal(1))

			Expect(fs.TouchCallCount()).To(Equal(3))
			Expect(fs.TouchArgsForCall(0)).To(Equal(filepath.Join(gitDir, pool, "claimed", ".gitkeep")))
			Expect(fs.TouchArgsForCall(1)).To(Equal(filepath.Join(gitDir, pool, "unclaimed", ".gitkeep")))
			Expect(fs.TouchArgsForCall(2)).To(Equal(filepath.Join(gitDir, pool, "unclaimed", pool)))

			message, actualUser := gitRepo.CommitAndPushArgsForCall(0)
			Expect(gitRepo.CommitAndPushCallCount()).To(Equal(1))
			Expect(message).To(Equal("Claimer creating " + pool))
			Expect(actualUser).To(Equal(user))
		})

		Context("when cloning the repo fails", func() {
			It("returns an error", func() {
				gitRepo.CloneOrPullReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.CreatePool("", "")).To(MatchError("failed to clone or pull: some-error"))
			})
		})

		Context("when touching claimed/.gitkeep fails", func() {
			It("returns an error", func() {
				fs.TouchReturnsOnCall(0, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.CreatePool("", "")).To(MatchError("failed to touch 'claimed/.gitkeep': some-error"))
			})
		})

		Context("when touching unclaimed/.gitkeep fails", func() {
			It("returns an error", func() {
				fs.TouchReturnsOnCall(1, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.CreatePool("", "")).To(MatchError("failed to touch 'unclaimed/.gitkeep': some-error"))
			})
		})

		Context("when touching the lock fails", func() {
			It("returns an error", func() {
				fs.TouchReturnsOnCall(2, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.CreatePool("", "")).To(MatchError("failed to touch lock file: some-error"))
			})
		})

		Context("when pushing fails", func() {
			It("returns an error", func() {
				gitRepo.CommitAndPushReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.CreatePool("", "")).To(MatchError("failed to commit and push: some-error"))
			})
		})
	})

	Describe("DestroyPool", func() {
		It("Destroys a pool", func() {
			pool := "some-pool"
			gitDir := "some-dir"
			user := "some-user"

			gitRepo.DirReturns(gitDir)

			locker := NewLocker(fs, gitRepo)
			Expect(locker.DestroyPool(pool, user)).To(Succeed())

			Expect(gitRepo.CloneOrPullCallCount()).To(Equal(1))

			Expect(fs.RmCallCount()).To(Equal(1))
			Expect(fs.RmArgsForCall(0)).To(Equal(filepath.Join(gitDir, pool)))

			message, actualUser := gitRepo.CommitAndPushArgsForCall(0)
			Expect(gitRepo.CommitAndPushCallCount()).To(Equal(1))
			Expect(message).To(Equal("Claimer destroying " + pool))
			Expect(actualUser).To(Equal(user))
		})

		Context("when cloning the repo fails", func() {
			It("returns an error", func() {
				gitRepo.CloneOrPullReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.DestroyPool("", "")).To(MatchError("failed to clone or pull: some-error"))
			})
		})

		Context("when removing the directory fails", func() {
			It("returns an error", func() {
				fs.RmReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.DestroyPool("", "")).To(MatchError("failed to remove directory: some-error"))
			})
		})

		Context("when pushing fails", func() {
			It("returns an error", func() {
				gitRepo.CommitAndPushReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.DestroyPool("", "")).To(MatchError("failed to commit and push: some-error"))
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
				Expect(locker.ReleaseLock("", "")).To(MatchError("failed to clone or pull: some-error"))
			})
		})

		Context("when listing files fails", func() {
			It("returns an error", func() {
				fs.LsReturns(nil, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ReleaseLock("", "")).To(MatchError("failed to list claimed locks: some-error"))
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
				Expect(locker.ReleaseLock("", "")).To(MatchError("failed to move file: some-error"))
			})
		})

		Context("when pushing fails", func() {
			It("returns an error", func() {
				fs.LsReturns([]string{"some-lock"}, nil)
				gitRepo.CommitAndPushReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				Expect(locker.ReleaseLock("", "")).To(MatchError("failed to commit and push: some-error"))
			})
		})
	})

	Describe("Status", func() {
		It("returns a list of locks", func() {
			author := "some-author"
			date := "some-date"
			message := "some-message"

			gitDir := "some-dir"
			gitRepo.DirReturns(gitDir)

			fs.LsDirsStub = func(dir string) ([]string, error) {
				if dir == gitDir {
					return []string{"pool-1", "pool-2", "pool-with-no-locks", "pool-with-many-locks"}, nil
				} else {
					return []string{}, nil
				}
			}
			fs.LsStub = func(dir string) ([]string, error) {
				if dir == filepath.Join(gitDir, "pool-1", "claimed") {
					return []string{"lock"}, nil
				} else if dir == filepath.Join(gitDir, "pool-2", "unclaimed") {
					return []string{"lock"}, nil
				} else if dir == filepath.Join(gitDir, "pool-with-many-locks", "claimed") {
					return []string{"lock-1"}, nil
				} else if dir == filepath.Join(gitDir, "pool-with-many-locks", "unclaimed") {
					return []string{"lock-2", "lock-3"}, nil
				} else {
					return []string{}, nil
				}
			}

			gitRepo.LatestCommitReturns(author, date, message, nil)

			locker := NewLocker(fs, gitRepo)
			locks, err := locker.Status()
			Expect(err).NotTo(HaveOccurred())
			Expect(locks).To(ConsistOf(
				Lock{Name: "pool-1", Claimed: true, Owner: author, Date: date, Message: message},
				Lock{Name: "pool-2", Claimed: false},
				Lock{Name: "pool-with-many-locks/lock-1", Claimed: true, Owner: author, Date: date, Message: message},
				Lock{Name: "pool-with-many-locks/lock-2", Claimed: false},
				Lock{Name: "pool-with-many-locks/lock-3", Claimed: false},
			))

			Expect(gitRepo.CloneOrPullCallCount()).To(Equal(1))
			Expect(gitRepo.LatestCommitCallCount()).To(Equal(2))
		})

		Context("when cloning the repo fails", func() {
			It("returns an error", func() {
				gitRepo.CloneOrPullReturns(errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				_, err := locker.Status()
				Expect(err).To(MatchError("failed to clone or pull: some-error"))
			})
		})

		Context("when listing the git repo fails", func() {
			It("returns an error", func() {
				fs.LsDirsReturns(nil, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				_, err := locker.Status()
				Expect(err).To(MatchError("failed to list pools: some-error"))
			})
		})

		Context("when listing claimed locks fails", func() {
			It("returns an error", func() {
				fs.LsDirsReturns([]string{"some-pool"}, nil)
				fs.LsReturns(nil, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				_, err := locker.Status()
				Expect(err).To(MatchError("failed to list claimed locks: some-error"))
			})
		})

		Context("when listing unclaimed locks fails", func() {
			It("returns an error", func() {
				fs.LsDirsReturns([]string{"some-pool"}, nil)
				fs.LsReturnsOnCall(1, nil, errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				_, err := locker.Status()
				Expect(err).To(MatchError("failed to list unclaimed locks: some-error"))
			})
		})

		Context("when getting the author fails", func() {
			It("returns an error", func() {
				fs.LsDirsReturns([]string{"some-pool"}, nil)
				fs.LsReturnsOnCall(0, []string{"some-lock"}, nil)
				gitRepo.LatestCommitReturns("", "", "", errors.New("some-error"))

				locker := NewLocker(fs, gitRepo)
				_, err := locker.Status()
				Expect(err).To(MatchError("failed to get latest commit: some-error"))
			})
		})
	})
})

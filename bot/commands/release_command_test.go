package commands_test

import (
	. "github.com/mdelillo/claimer/bot/commands"

	"errors"
	"github.com/mdelillo/claimer/bot/commands/commandsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseCommand", func() {
	Describe("Execute", func() {
		var locker *commandsfakes.FakeLocker

		BeforeEach(func() {
			locker = new(commandsfakes.FakeLocker)
		})

		It("releases the lock and returns a slack response", func() {
			pool := "some-pool"
			username := "some-username"

			locker.StatusReturns([]string{pool}, []string{}, nil)

			command := NewFactory(locker).NewCommand("release", pool, username)

			slackResponse, err := command.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(slackResponse).To(Equal("Released " + pool))

			Expect(locker.ReleaseLockCallCount()).To(Equal(1))
			actualPool, actualUsername := locker.ReleaseLockArgsForCall(0)
			Expect(actualPool).To(Equal(pool))
			Expect(actualUsername).To(Equal(username))
		})

		Context("when no pool is specified", func() {
			It("returns an error", func() {
				command := NewFactory(locker).NewCommand("release", "", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("no pool specified"))
				Expect(slackResponse).To(BeEmpty())
			})
		})

		Context("when the pool is not claimed", func() {
			It("returns a slack response", func() {
				pool := "some-pool"

				locker.StatusReturns([]string{}, []string{}, nil)

				command := NewFactory(locker).NewCommand("release", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(pool + " is not claimed"))
			})
		})

		Context("when checking the status fails", func() {
			It("returns an error", func() {
				locker.StatusReturns(nil, nil, errors.New("some-error"))

				command := NewFactory(locker).NewCommand("release", "some-pool", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to get status of locks: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})

		Context("when releasing the lock fails", func() {
			It("returns an error", func() {
				pool := "some-pool"

				locker.StatusReturns([]string{pool}, []string{}, nil)
				locker.ReleaseLockReturns(errors.New("some-error"))

				command := NewFactory(locker).NewCommand("release", "some-pool", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to release lock: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})
	})
})

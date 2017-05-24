package commands_test

import (
	. "github.com/mdelillo/claimer/bot/commands"

	"errors"
	"github.com/mdelillo/claimer/bot/commands/commandsfakes"
	clocker "github.com/mdelillo/claimer/locker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClaimCommand", func() {
	Describe("Execute", func() {
		var locker *commandsfakes.FakeLocker

		BeforeEach(func() {
			locker = new(commandsfakes.FakeLocker)
		})

		Context("when no message is provided", func() {
			It("claims the lock and returns a slack response", func() {
				pool := "some-pool"
				username := "some-username"

				locker.StatusReturns(
					[]clocker.Lock{{Name: pool, Claimed: false}},
					nil,
				)

				command := NewFactory(locker).NewCommand("claim", pool, username)

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal("Claimed " + pool))

				Expect(locker.ClaimLockCallCount()).To(Equal(1))
				actualPool, actualUsername, actualMessage := locker.ClaimLockArgsForCall(0)
				Expect(actualPool).To(Equal(pool))
				Expect(actualUsername).To(Equal(username))
				Expect(actualMessage).To(BeEmpty())
			})
		})

		Context("when a message is provided", func() {
			It("claims the lock passing along the message", func() {
				pool := "some-pool"
				message := "some message"
				username := "some-username"

				locker.StatusReturns(
					[]clocker.Lock{{Name: pool, Claimed: false}},
					nil,
				)

				command := NewFactory(locker).NewCommand("claim", pool+" "+message, username)

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal("Claimed " + pool))

				Expect(locker.ClaimLockCallCount()).To(Equal(1))
				actualPool, actualUsername, actualMessage := locker.ClaimLockArgsForCall(0)
				Expect(actualPool).To(Equal(pool))
				Expect(actualUsername).To(Equal(username))
				Expect(actualMessage).To(Equal(message))
			})
		})

		Context("when no pool is specified", func() {
			It("returns an error", func() {
				command := NewFactory(locker).NewCommand("claim", "", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("no pool specified"))
				Expect(slackResponse).To(BeEmpty())
			})
		})

		Context("when the pool does not exist", func() {
			It("returns a slack response", func() {
				pool := "some-pool"

				locker.StatusReturns(nil, nil)

				command := NewFactory(locker).NewCommand("claim", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(pool + " does not exist"))
			})
		})

		Context("when the pool is already claimed", func() {
			It("returns a slack response", func() {
				pool := "some-pool"

				locker.StatusReturns(
					[]clocker.Lock{{Name: pool, Claimed: true}},
					nil,
				)

				command := NewFactory(locker).NewCommand("claim", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(pool + " is already claimed"))
			})
		})

		Context("when checking the status fails", func() {
			It("returns an error", func() {
				locker.StatusReturns(nil, errors.New("some-error"))

				command := NewFactory(locker).NewCommand("claim", "some-pool", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to get status of locks: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})

		Context("when claiming the lock fails", func() {
			It("returns an error", func() {
				pool := "some-pool"

				locker.StatusReturns(
					[]clocker.Lock{{Name: pool, Claimed: false}},
					nil,
				)
				locker.ClaimLockReturns(errors.New("some-error"))

				command := NewFactory(locker).NewCommand("claim", "some-pool", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to claim lock: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})
	})
})

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
				lock := "some-lock"
				username := "some-username"

				locker.StatusReturns(
					[]clocker.Lock{{Name: lock, Claimed: false}},
					nil,
				)

				command := NewFactory(locker).NewCommand("claim", lock, username)

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal("Claimed " + lock))

				Expect(locker.ClaimLockCallCount()).To(Equal(1))
				actualLock, actualUsername, actualMessage := locker.ClaimLockArgsForCall(0)
				Expect(actualLock).To(Equal(lock))
				Expect(actualUsername).To(Equal(username))
				Expect(actualMessage).To(BeEmpty())
			})
		})

		Context("when a message is provided", func() {
			It("claims the lock passing along the message", func() {
				lock := "some-lock"
				message := "some message"
				username := "some-username"

				locker.StatusReturns(
					[]clocker.Lock{{Name: lock, Claimed: false}},
					nil,
				)

				command := NewFactory(locker).NewCommand("claim", lock+" "+message, username)

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal("Claimed " + lock))

				Expect(locker.ClaimLockCallCount()).To(Equal(1))
				actualLock, actualUsername, actualMessage := locker.ClaimLockArgsForCall(0)
				Expect(actualLock).To(Equal(lock))
				Expect(actualUsername).To(Equal(username))
				Expect(actualMessage).To(Equal(message))
			})
		})

		Context("when no lock is specified", func() {
			It("returns a slack response", func() {
				command := NewFactory(locker).NewCommand("claim", "", "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal("must specify lock to claim"))
			})
		})

		Context("when the lock does not exist", func() {
			It("returns a slack response", func() {
				lock := "some-lock"

				locker.StatusReturns(nil, nil)

				command := NewFactory(locker).NewCommand("claim", lock, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(lock + " does not exist"))
			})
		})

		Context("when the lock is already claimed", func() {
			It("returns a slack response", func() {
				lock := "some-lock"

				locker.StatusReturns(
					[]clocker.Lock{{Name: lock, Claimed: true}},
					nil,
				)

				command := NewFactory(locker).NewCommand("claim", lock, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(lock + " is already claimed"))
			})
		})

		Context("when checking the status fails", func() {
			It("returns an error", func() {
				locker.StatusReturns(nil, errors.New("some-error"))

				command := NewFactory(locker).NewCommand("claim", "some-lock", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to get status of locks: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})

		Context("when claiming the lock fails", func() {
			It("returns an error", func() {
				lock := "some-lock"

				locker.StatusReturns(
					[]clocker.Lock{{Name: lock, Claimed: false}},
					nil,
				)
				locker.ClaimLockReturns(errors.New("some-error"))

				command := NewFactory(locker).NewCommand("claim", "some-lock", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to claim lock: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})
	})
})

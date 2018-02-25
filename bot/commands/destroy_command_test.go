package commands_test

import (
	. "github.com/mdelillo/claimer/bot/commands"

	"errors"

	"github.com/mdelillo/claimer/bot/commands/commandsfakes"
	clocker "github.com/mdelillo/claimer/locker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DestroyCommand", func() {
	Describe("Execute", func() {
		var locker *commandsfakes.FakeLocker

		BeforeEach(func() {
			locker = new(commandsfakes.FakeLocker)
		})

		It("destroys the pool and returns a slack response", func() {
			pool := "some-pool"
			username := "some-username"

			locker.StatusReturns(
				[]clocker.Lock{{Name: pool, Claimed: true}},
				nil,
			)

			command := NewFactory(locker).NewCommand("destroy", pool, username)

			slackResponse, err := command.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(slackResponse).To(Equal("Destroyed " + pool))

			Expect(locker.DestroyPoolCallCount()).To(Equal(1))
			actualPool, actualUsername := locker.DestroyPoolArgsForCall(0)
			Expect(actualPool).To(Equal(pool))
			Expect(actualUsername).To(Equal(username))
		})

		Context("when no pool is specified", func() {
			It("returns a slack response", func() {
				command := NewFactory(locker).NewCommand("destroy", "", "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal("must specify pool to destroy"))
			})
		})

		Context("when checking the status fails", func() {
			It("returns an error", func() {
				locker.StatusReturns(nil, errors.New("some-error"))

				command := NewFactory(locker).NewCommand("destroy", "some-pool", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to get status of locks: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})

		Context("when the pool does not exist", func() {
			It("returns a slack response", func() {
				pool := "some-pool"

				locker.StatusReturns(nil, nil)

				command := NewFactory(locker).NewCommand("destroy", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(pool + " does not exist"))
			})
		})

		Context("when destroying the pool fails", func() {
			It("returns an error", func() {
				pool := "some-pool"

				locker.StatusReturns(
					[]clocker.Lock{{Name: pool, Claimed: false}},
					nil,
				)
				locker.DestroyPoolReturns(errors.New("some-error"))

				command := NewFactory(locker).NewCommand("destroy", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to destroy pool: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})
	})
})

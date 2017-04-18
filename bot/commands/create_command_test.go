package commands_test

import (
	. "github.com/mdelillo/claimer/bot/commands"

	"errors"
	"github.com/mdelillo/claimer/bot/commands/commandsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateCommand", func() {
	Describe("Execute", func() {
		var locker *commandsfakes.FakeLocker

		BeforeEach(func() {
			locker = new(commandsfakes.FakeLocker)
		})

		It("creates the pool and returns a slack response", func() {
			pool := "some-pool"
			username := "some-username"

			locker.StatusReturns([]string{}, []string{}, nil)

			command := NewFactory(locker).NewCommand("create", pool, username)

			slackResponse, err := command.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(slackResponse).To(Equal("Created " + pool))

			Expect(locker.CreatePoolCallCount()).To(Equal(1))
			actualPool, actualUsername := locker.CreatePoolArgsForCall(0)
			Expect(actualPool).To(Equal(pool))
			Expect(actualUsername).To(Equal(username))
		})

		Context("when no pool is specified", func() {
			It("returns an error", func() {
				command := NewFactory(locker).NewCommand("create", "", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("no pool specified"))
				Expect(slackResponse).To(BeEmpty())
			})
		})

		Context("when checking the status fails", func() {
			It("returns an error", func() {
				locker.StatusReturns(nil, nil, errors.New("some-error"))

				command := NewFactory(locker).NewCommand("create", "some-pool", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to get status of locks: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})

		Context("when the pool already exists and is claimed", func() {
			It("returns a slack response", func() {
				pool := "some-pool"

				locker.StatusReturns([]string{pool}, []string{}, nil)

				command := NewFactory(locker).NewCommand("create", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(pool + " already exists"))
			})
		})

		Context("when the pool already exists and is unclaimed", func() {
			It("returns a slack response", func() {
				pool := "some-pool"

				locker.StatusReturns([]string{}, []string{pool}, nil)

				command := NewFactory(locker).NewCommand("create", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(pool + " already exists"))
			})
		})

		Context("when creating the pool fails", func() {
			It("returns an error", func() {
				locker.CreatePoolReturns(errors.New("some-error"))

				command := NewFactory(locker).NewCommand("create", "some-pool", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to create pool: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})
	})
})

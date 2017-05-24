package commands_test

import (
	. "github.com/mdelillo/claimer/bot/commands"

	"errors"
	"github.com/mdelillo/claimer/bot/commands/commandsfakes"
	clocker "github.com/mdelillo/claimer/locker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatusCommand", func() {
	Describe("Execute", func() {
		var locker *commandsfakes.FakeLocker

		BeforeEach(func() {
			locker = new(commandsfakes.FakeLocker)
		})

		It("responds with the status of the locks", func() {
			username := "some-username"
			locker.StatusReturns(
				[]clocker.Lock{
					{Name: "claimed-1", Owner: username, Claimed: true},
					{Name: "claimed-2", Owner: "some-other-user", Claimed: true},
					{Name: "unclaimed-1", Claimed: false},
					{Name: "unclaimed-2", Claimed: false},
				},
				nil,
			)

			command := NewFactory(locker).NewCommand("status", "", username)

			slackResponse, err := command.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(slackResponse).To(Equal("*Claimed by you:* claimed-1\n*Claimed by others:* claimed-2\n*Unclaimed:* unclaimed-1, unclaimed-2"))
		})

		Context("when getting the status fails", func() {
			It("returns an error", func() {
				locker.StatusReturns(nil, errors.New("some-error"))

				command := NewFactory(locker).NewCommand("status", "", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to get status of locks: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})
	})
})

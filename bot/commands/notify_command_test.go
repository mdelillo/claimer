package commands_test

import (
	"errors"

	. "github.com/mdelillo/claimer/bot/commands"
	"github.com/mdelillo/claimer/bot/commands/commandsfakes"
	clocker "github.com/mdelillo/claimer/locker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NotifyCommand", func() {
	Describe("Execute", func() {
		var locker *commandsfakes.FakeLocker

		BeforeEach(func() {
			locker = new(commandsfakes.FakeLocker)
		})

		Context("when there are no claimed locks", func() {
			It("gives an informative message", func() {
				locker.StatusReturns(
					[]clocker.Lock{
						{Name: "unclaimed-1", Claimed: false},
						{Name: "unclaimed-2", Claimed: false},
					},
					nil,
				)

				command := NewFactory(locker).NewCommand("notify", "", "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal("No locks currently claimed."))
			})
		})

		It("responds with owners and their claimed locks", func() {
			username := "some-username"
			locker.StatusReturns(
				[]clocker.Lock{
					{Name: "claimed-1", Owner: username, Claimed: true},
					{Name: "claimed-2", Owner: "some-other-user", Claimed: true},
					{Name: "claimed-3", Owner: username, Claimed: true},
					{Name: "unclaimed-1", Claimed: false},
					{Name: "unclaimed-2", Claimed: false},
				},
				nil,
			)

			command := NewFactory(locker).NewCommand("notify", "", username)

			slackResponse, err := command.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(slackResponse).To(ContainSubstring("Currently claimed locks, please release if not in use:"))
			Expect(slackResponse).To(ContainSubstring("<@some-username>: claimed-1, claimed-3"))
			Expect(slackResponse).To(ContainSubstring("<@some-other-user>: claimed-2"))
		})

		Context("when notifying fails", func() {
			It("returns an error", func() {
				locker.StatusReturns(nil, errors.New("some-error"))

				command := NewFactory(locker).NewCommand("notify", "", "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to get status of locks: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})
	})
})

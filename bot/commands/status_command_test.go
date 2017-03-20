package commands_test

import (
	. "github.com/mdelillo/claimer/bot/commands"

	"errors"
	"github.com/mdelillo/claimer/bot/commands/commandsfakes"
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
			locker.StatusReturns(
				[]string{"claimed-1", "claimed-2"},
				[]string{"unclaimed-1", "unclaimed-2"},
				nil,
			)

			command := NewFactory(locker).NewCommand("status", []string{}, "")

			slackResponse, err := command.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(slackResponse).To(Equal("*Claimed:* claimed-1, claimed-2\n*Unclaimed:* unclaimed-1, unclaimed-2"))
		})

		Context("when getting the status fails", func() {
			It("returns an error", func() {
				locker.StatusReturns(nil, nil, errors.New("some-error"))

				command := NewFactory(locker).NewCommand("status", []string{}, "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to get status of locks: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})
	})
})

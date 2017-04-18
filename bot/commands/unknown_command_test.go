package commands_test

import (
	. "github.com/mdelillo/claimer/bot/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnknownCommand", func() {
	Describe("Execute", func() {
		It("returns a slack response", func() {
			command := NewFactory(nil).NewCommand("some-bad-command", "", "")

			slackResponse, err := command.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(slackResponse).To(Equal("Unknown command. Try `@claimer help` to see usage."))
		})
	})
})

package commands_test

import (
	. "github.com/mdelillo/claimer/bot/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HelpCommand", func() {
	Describe("Execute", func() {
		It("returns the help text", func() {
			command := NewFactory(nil).NewCommand("help", "", "")

			slackResponse, err := command.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(slackResponse).To(Equal(
				"Available commands:\n" +
					"```\n" +
					"  claim <env> [<message>]   Claim an unclaimed environment\n" +
					"  create <env>              Create a new environment\n" +
					"  destroy <env>             Destroy an environment\n" +
					"  owner <env>               Show the user who claimed the environment\n" +
					"  release <env>             Release a claimed environment\n" +
					"  status                    Show claimed and unclaimed environments\n" +
					"  help                      Display this message\n" +
					"```",
			))
		})
	})
})

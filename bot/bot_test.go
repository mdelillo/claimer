package bot_test

import (
	. "github.com/mdelillo/claimer/bot"

	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	logrustest "github.com/Sirupsen/logrus/hooks/test"
	"github.com/mdelillo/claimer/bot/botfakes"
	"github.com/mdelillo/claimer/bot/commands/commandsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bot", func() {
	Describe("Run", func() {
		var (
			command        *commandsfakes.FakeCommand
			commandFactory *botfakes.FakeCommandFactory
			slackClient    *botfakes.FakeSlackClient
			logger         *logrus.Logger
			logHook        *logrustest.Hook
		)

		BeforeEach(func() {
			command = new(commandsfakes.FakeCommand)
			commandFactory = new(botfakes.FakeCommandFactory)
			slackClient = new(botfakes.FakeSlackClient)
			logger, logHook = logrustest.NewNullLogger()
		})

		AfterEach(func() {
			logHook.Reset()
		})

		Context("when arguments are provided", func() {
			It("executes a command with arguments and posts the response in slack", func() {
				cmd := "some-command"
				args := "some-arg some-other-arg"
				channel := "some-channel"
				username := "some-username"
				message := "some-message"

				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler(fmt.Sprintf("<@some-bot> %s %s", cmd, args), channel, username)
					return nil
				}
				commandFactory.NewCommandReturns(command)
				command.ExecuteReturns(message, nil)

				Expect(New(commandFactory, slackClient, logger).Run()).To(Succeed())

				actualCmd, actualArgs, actualUsername := commandFactory.NewCommandArgsForCall(0)
				Expect(actualCmd).To(Equal(cmd))
				Expect(actualArgs).To(Equal(args))
				Expect(actualUsername).To(Equal(username))

				actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(actualChannel).To(Equal(channel))
				Expect(actualMessage).To(Equal(message))
			})
		})

		Context("when no arguments are provided", func() {
			It("executes a command and posts the response in slack", func() {
				cmd := "some-command"
				channel := "some-channel"
				username := "some-username"
				message := "some-message"

				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler(fmt.Sprintf("<@some-bot> %s", cmd), channel, username)
					return nil
				}
				commandFactory.NewCommandReturns(command)
				command.ExecuteReturns(message, nil)

				Expect(New(commandFactory, slackClient, logger).Run()).To(Succeed())

				actualCmd, actualArgs, actualUsername := commandFactory.NewCommandArgsForCall(0)
				Expect(actualCmd).To(Equal(cmd))
				Expect(actualArgs).To(BeEmpty())
				Expect(actualUsername).To(Equal(username))

				actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(actualChannel).To(Equal(channel))
				Expect(actualMessage).To(Equal(message))
			})
		})

		Context("when text precedes call to claimer", func() {
			It("executes a command with arguments and posts the response in slack", func() {
				cmd := "some-command"
				args := "some-arg some-other-arg"
				channel := "some-channel"
				username := "some-username"
				message := "some-message"

				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler(fmt.Sprintf("something <@some-bot> %s %s", cmd, args), channel, username)
					return nil
				}
				commandFactory.NewCommandReturns(command)
				command.ExecuteReturns(message, nil)

				Expect(New(commandFactory, slackClient, logger).Run()).To(Succeed())

				actualCmd, actualArgs, actualUsername := commandFactory.NewCommandArgsForCall(0)
				Expect(actualCmd).To(Equal(cmd))
				Expect(actualArgs).To(Equal(args))
				Expect(actualUsername).To(Equal(username))

				actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(actualChannel).To(Equal(channel))
				Expect(actualMessage).To(Equal(message))
			})
		})

		Context("when no command is specified", func() {
			It("uses an empty value for the command", func() {
				channel := "some-channel"
				username := "some-username"
				message := "some-message"

				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler("<@some-bot>", channel, username)
					return nil
				}
				commandFactory.NewCommandReturns(command)
				command.ExecuteReturns(message, nil)

				Expect(New(commandFactory, slackClient, logger).Run()).To(Succeed())

				actualCmd, actualArgs, actualUsername := commandFactory.NewCommandArgsForCall(0)
				Expect(actualCmd).To(BeEmpty())
				Expect(actualArgs).To(BeEmpty())
				Expect(actualUsername).To(Equal(username))

				actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(actualChannel).To(Equal(channel))
				Expect(actualMessage).To(Equal(message))
			})
		})

		Context("when listening fails", func() {
			It("returns an error", func() {
				slackClient.ListenReturns(errors.New("some-error"))
				Expect(New(commandFactory, slackClient, logger).Run()).To(MatchError("some-error"))
			})
		})

		Context("when the command returns an error", func() {
			It("logs an error", func() {
				text := "<@some-bot> some-command"
				channel := "some-channel"
				username := "some-username"

				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler(text, channel, username)
					return nil
				}
				commandFactory.NewCommandReturns(command)
				command.ExecuteReturns("", errors.New("some-error"))

				Expect(New(commandFactory, slackClient, logger).Run()).To(Succeed())

				Expect(len(logHook.Entries)).To(Equal(1))
				Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
				Expect(logHook.LastEntry().Message).To(Equal("failed to execute command"))
				Expect(logHook.LastEntry().Data["error"]).To(Equal("some-error"))
				Expect(logHook.LastEntry().Data["text"]).To(Equal(text))
				Expect(logHook.LastEntry().Data["channel"]).To(Equal(channel))
				Expect(logHook.LastEntry().Data["username"]).To(Equal(username))

				Expect(slackClient.PostMessageCallCount()).To(Equal(0))
			})
		})

		Context("when posting to slack fails", func() {
			It("logs an error", func() {
				text := "<@some-bot> some-command"
				channel := "some-channel"
				username := "some-username"

				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler(text, channel, username)
					return nil
				}
				commandFactory.NewCommandReturns(command)
				command.ExecuteReturns("some-response", nil)
				slackClient.PostMessageReturns(errors.New("some-error"))

				Expect(New(commandFactory, slackClient, logger).Run()).To(Succeed())

				Expect(len(logHook.Entries)).To(Equal(1))
				Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
				Expect(logHook.LastEntry().Message).To(Equal("failed to post to slack: some-error"))
			})
		})
	})
})

package bot_test

import (
	. "github.com/mdelillo/claimer/bot"

	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	logrustest "github.com/Sirupsen/logrus/hooks/test"
	"github.com/mdelillo/claimer/bot/botfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bot", func() {
	Describe("Run", func() {
		var (
			locker      *botfakes.FakeLocker
			slackClient *botfakes.FakeSlackClient
			logger      *logrus.Logger
			logHook     *logrustest.Hook
			channel     = "some-channel"
			pool        = "some-pool"
			username    = "some-username"
		)

		BeforeEach(func() {
			locker = new(botfakes.FakeLocker)
			slackClient = new(botfakes.FakeSlackClient)
			logger, logHook = logrustest.NewNullLogger()
		})

		AfterEach(func() {
			logHook.Reset()
		})

		Context("when a claim command is received", func() {
			It("claims the lock and posts to slack", func() {
				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler(fmt.Sprintf("@some-bot claim %s", pool), channel, username)
					return nil
				}
				locker.StatusReturns([]string{}, []string{pool}, nil)

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())

				Expect(locker.ClaimLockCallCount()).To(Equal(1))
				actualPool, actualUsername := locker.ClaimLockArgsForCall(0)
				Expect(actualPool).To(Equal(pool))
				Expect(actualUsername).To(Equal(username))

				actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(actualChannel).To(Equal(channel))
				Expect(actualMessage).To(Equal("Claimed " + pool))
			})

			Context("when no pool is specified", func() {
				It("logs an error", func() {
					text := "@some-bot claim"
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "no pool specified", text, channel, username)
				})
			})

			Context("when the pool is not available", func() {
				It("responds in slack", func() {
					text := fmt.Sprintf("@some-bot claim %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns([]string{}, []string{}, nil)

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(locker.ClaimLockCallCount()).To(Equal(0))

					actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
					Expect(slackClient.PostMessageCallCount()).To(Equal(1))
					Expect(actualChannel).To(Equal(channel))
					Expect(actualMessage).To(Equal(pool + " is not available"))
				})
			})

			Context("when checking the status fails", func() {
				It("logs an error", func() {
					text := "@some-bot claim some-pool"
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns(nil, nil, errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel, username)
				})
			})

			Context("when posting error to slack fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot claim %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns([]string{}, []string{}, nil)
					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(len(logHook.Entries)).To(Equal(1))
					Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
					Expect(logHook.LastEntry().Message).To(Equal("failed to post to slack: some-error"))
				})
			})

			Context("when claiming the lock fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot claim %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns([]string{}, []string{pool}, nil)
					locker.ClaimLockReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel, username)
				})
			})

			Context("when posting success to slack fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot claim %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns([]string{}, []string{pool}, nil)
					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(len(logHook.Entries)).To(Equal(1))
					Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
					Expect(logHook.LastEntry().Message).To(Equal("failed to post to slack: some-error"))
				})
			})
		})

		Context("when a help command is received", func() {
			It("posts the help message to slack", func() {
				text := "@some-bot help"
				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler(text, channel, username)
					return nil
				}

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())

				actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(actualChannel).To(Equal(channel))
				Expect(actualMessage).To(Equal(
					"Available commands:\n" +
						"```\n" +
						"  claim <env>     Claim an unclaimed environment\n" +
						"  owner <env>     Show the user who claimed the environment\n" +
						"  release <env>   Release a claimed environment\n" +
						"  status          Show claimed and unclaimed environments\n" +
						"  help            Display this message\n" +
						"```",
				))
			})

			Context("when posting to slack fails", func() {
				It("logs an error", func() {
					text := "@some-bot help"
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}

					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(len(logHook.Entries)).To(Equal(1))
					Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
					Expect(logHook.LastEntry().Message).To(Equal("failed to post to slack: some-error"))
				})
			})
		})

		Context("when an owner command is received", func() {
			Context("when the lock is claimed", func() {
				It("posts the owner of the lock to slack", func() {
					claimDate := "some-date"

					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(fmt.Sprintf("@some-bot owner %s", pool), channel, username)
						return nil
					}
					locker.StatusReturns([]string{pool}, []string{}, nil)
					locker.OwnerReturns(username, claimDate, nil)

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(locker.OwnerCallCount()).To(Equal(1))
					Expect(locker.OwnerArgsForCall(0)).To(Equal(pool))

					actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
					Expect(slackClient.PostMessageCallCount()).To(Equal(1))
					Expect(actualChannel).To(Equal(channel))
					Expect(actualMessage).To(Equal(fmt.Sprintf("%s was claimed by %s on %s", pool, username, claimDate)))
				})
			})

			Context("when the lock is not claimed", func() {
				It("posts in slack", func() {
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(fmt.Sprintf("@some-bot owner %s", pool), channel, username)
						return nil
					}
					locker.StatusReturns([]string{}, []string{}, nil)

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
					Expect(slackClient.PostMessageCallCount()).To(Equal(1))
					Expect(actualChannel).To(Equal(channel))
					Expect(actualMessage).To(Equal(pool + " is not claimed"))
				})
			})

			Context("when no pool is specified", func() {
				It("logs an error", func() {
					text := "@some-bot owner"
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "no pool specified", text, channel, username)
				})
			})

			Context("when posting unclaimed in slack fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot owner %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns([]string{}, []string{}, nil)
					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(len(logHook.Entries)).To(Equal(1))
					Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
					Expect(logHook.LastEntry().Message).To(Equal("failed to post to slack: some-error"))
				})
			})

			Context("when checking the owner fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot owner %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns([]string{pool}, []string{}, nil)
					locker.OwnerReturns("", "", errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel, username)
				})
			})

			Context("when posting owner to slack fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot owner %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns([]string{pool}, []string{}, nil)
					locker.OwnerReturns("", "", nil)
					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(len(logHook.Entries)).To(Equal(1))
					Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
					Expect(logHook.LastEntry().Message).To(Equal("failed to post to slack: some-error"))
				})
			})
		})

		Context("when a release command is received", func() {
			It("releases the lock and posts to slack", func() {
				text := fmt.Sprintf("@some-bot release %s", pool)
				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler(text, channel, username)
					return nil
				}
				locker.StatusReturns([]string{pool}, []string{}, nil)

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())

				Expect(locker.ReleaseLockCallCount()).To(Equal(1))
				actualPool, actualUsername := locker.ReleaseLockArgsForCall(0)
				Expect(actualPool).To(Equal(pool))
				Expect(actualUsername).To(Equal(username))

				actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(actualChannel).To(Equal(channel))
				Expect(actualMessage).To(Equal("Released " + pool))
			})

			Context("when no pool is specified", func() {
				It("logs an error", func() {
					text := "@some-bot release"
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "no pool specified", text, channel, username)
				})
			})

			Context("when the pool is not claimed", func() {
				It("responds in slack", func() {
					text := fmt.Sprintf("@some-bot release %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns([]string{}, []string{}, nil)

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(locker.ReleaseLockCallCount()).To(Equal(0))

					actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
					Expect(slackClient.PostMessageCallCount()).To(Equal(1))
					Expect(actualChannel).To(Equal(channel))
					Expect(actualMessage).To(Equal(pool + " is not claimed"))
				})
			})

			Context("when checking the status fails", func() {
				It("logs an error", func() {
					text := "@some-bot release some-pool"
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns(nil, nil, errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel, username)
				})
			})

			Context("when posting error to slack fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot release %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns([]string{}, []string{}, nil)
					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(len(logHook.Entries)).To(Equal(1))
					Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
					Expect(logHook.LastEntry().Message).To(Equal("failed to post to slack: some-error"))
				})
			})

			Context("when releasing the lock fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot release %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns([]string{pool}, []string{}, nil)

					locker.ReleaseLockReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel, username)
				})
			})

			Context("when posting to slack fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot release %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}
					locker.StatusReturns([]string{pool}, []string{}, nil)

					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(len(logHook.Entries)).To(Equal(1))
					Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
					Expect(logHook.LastEntry().Message).To(Equal("failed to post to slack: some-error"))
				})
			})
		})

		Context("when a status command is received", func() {
			It("posts the status of locks to slack", func() {
				text := "@some-bot status"
				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler(text, channel, username)
					return nil
				}

				locker.StatusReturns(
					[]string{"claimed-1", "claimed-2"},
					[]string{"unclaimed-1", "unclaimed-2"},
					nil,
				)

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())

				Expect(locker.StatusCallCount()).To(Equal(1))

				actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(actualChannel).To(Equal(channel))
				Expect(actualMessage).To(Equal("*Claimed:* claimed-1, claimed-2\n*Unclaimed:* unclaimed-1, unclaimed-2"))
			})

			Context("when getting the status fails", func() {
				It("logs an error", func() {
					text := "@some-bot status"
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}

					locker.StatusReturns(nil, nil, errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel, username)
				})
			})

			Context("when posting to slack fails", func() {
				It("logs an error", func() {
					text := "@some-bot status"
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}

					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(len(logHook.Entries)).To(Equal(1))
					Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
					Expect(logHook.LastEntry().Message).To(Equal("failed to post to slack: some-error"))
				})
			})
		})

		Context("when listening to slack messages fails", func() {
			It("returns an error", func() {
				slackClient.ListenReturns(errors.New("some-error"))
				claimer := New(locker, slackClient, logger)
				Expect(claimer.Run()).To(MatchError("some-error"))
			})
		})

		Context("when no command is specified", func() {
			It("responds in slack", func() {
				text := "@some-bot"
				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler(text, channel, username)
					return nil
				}

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())

				actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(actualChannel).To(Equal(channel))
				Expect(actualMessage).To(Equal("No command specified. Try `@claimer help` to see usage."))
			})

			Context("when posting to slack fails", func() {
				It("logs an error", func() {
					text := "@some-bot"
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}

					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(len(logHook.Entries)).To(Equal(1))
					Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
					Expect(logHook.LastEntry().Message).To(Equal("failed to post to slack: some-error"))
				})

			})
		})

		Context("when an unknown command is received", func() {
			It("responds in slack", func() {
				text := "@some-bot some-bad-command"
				slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
					messageHandler(text, channel, username)
					return nil
				}

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())

				actualChannel, actualMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(actualChannel).To(Equal(channel))
				Expect(actualMessage).To(Equal("Unknown command. Try `@claimer help` to see usage."))
			})

			Context("when posting to slack fails", func() {
				It("logs an error", func() {
					text := "@some-bot some-bad-command"
					slackClient.ListenStub = func(messageHandler func(_, _, _ string)) error {
						messageHandler(text, channel, username)
						return nil
					}

					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(len(logHook.Entries)).To(Equal(1))
					Expect(logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
					Expect(logHook.LastEntry().Message).To(Equal("failed to post to slack: some-error"))
				})

			})
		})
	})
})

func expectLoggedMessageHandlingError(logHook *logrustest.Hook, err, text, channel, username string) {
	ExpectWithOffset(1, len(logHook.Entries)).To(Equal(1))
	ExpectWithOffset(1, logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
	ExpectWithOffset(1, logHook.LastEntry().Message).To(Equal("Failed to handle message"))
	ExpectWithOffset(1, logHook.LastEntry().Data["error"]).To(Equal(err))
	ExpectWithOffset(1, logHook.LastEntry().Data["text"]).To(Equal(text))
	ExpectWithOffset(1, logHook.LastEntry().Data["channel"]).To(Equal(channel))
	ExpectWithOffset(1, logHook.LastEntry().Data["username"]).To(Equal(username))
}

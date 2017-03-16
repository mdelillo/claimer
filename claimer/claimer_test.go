package claimer_test

import (
	. "github.com/mdelillo/claimer/claimer"

	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	logrustest "github.com/Sirupsen/logrus/hooks/test"
	"github.com/mdelillo/claimer/claimer/claimerfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Claimer", func() {
	Describe("Run", func() {
		var (
			locker      *claimerfakes.FakeLocker
			slackClient *claimerfakes.FakeSlackClient
			logger      *logrus.Logger
			logHook     *logrustest.Hook
			channel     = "some-channel"
			pool        = "some-pool"
		)

		BeforeEach(func() {
			locker = new(claimerfakes.FakeLocker)
			slackClient = new(claimerfakes.FakeSlackClient)
			logger, logHook = logrustest.NewNullLogger()
		})

		AfterEach(func() {
			logHook.Reset()
		})

		Context("when a claim command is received", func() {
			It("claims the lock and posts to slack", func() {
				slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
					messageHandler(fmt.Sprintf("@some-bot claim %s", pool), channel)
					return nil
				}
				locker.StatusReturns([]string{}, []string{pool}, nil)

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())

				Expect(locker.ClaimLockCallCount()).To(Equal(1))
				Expect(locker.ClaimLockArgsForCall(0)).To(Equal(pool))

				postChannel, postMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(postChannel).To(Equal(channel))
				Expect(postMessage).To(Equal("Claimed " + pool))
			})

			Context("when no pool is specified", func() {
				It("logs an error", func() {
					text := "@some-bot claim"
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "no pool specified", text, channel)
				})
			})

			Context("when the pool is not available", func() {
				It("responds in slack", func() {
					text := fmt.Sprintf("@some-bot claim %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}
					locker.StatusReturns([]string{}, []string{}, nil)

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(locker.ClaimLockCallCount()).To(Equal(0))

					postChannel, postMessage := slackClient.PostMessageArgsForCall(0)
					Expect(slackClient.PostMessageCallCount()).To(Equal(1))
					Expect(postChannel).To(Equal(channel))
					Expect(postMessage).To(Equal(pool + " is not available"))
				})
			})

			Context("when checking the status fails", func() {
				It("logs an error", func() {
					text := "@some-bot claim some-pool"
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}
					locker.StatusReturns(nil, nil, errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel)
				})
			})

			Context("when posting error to slack fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot claim %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}
					locker.StatusReturns([]string{}, []string{}, nil)
					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel)
				})
			})

			Context("when claiming the lock fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot claim %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}
					locker.StatusReturns([]string{}, []string{pool}, nil)
					locker.ClaimLockReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel)
				})
			})

			Context("when posting success to slack fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot claim %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}
					locker.StatusReturns([]string{}, []string{pool}, nil)
					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel)
				})
			})
		})

		Context("when a release command is received", func() {
			It("releases the lock and posts to slack", func() {
				text := fmt.Sprintf("@some-bot release %s", pool)
				slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
					messageHandler(text, channel)
					return nil
				}
				locker.StatusReturns([]string{pool}, []string{}, nil)

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())

				Expect(locker.ReleaseLockCallCount()).To(Equal(1))
				Expect(locker.ReleaseLockArgsForCall(0)).To(Equal(pool))

				postChannel, postMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(postChannel).To(Equal(channel))
				Expect(postMessage).To(Equal("Released " + pool))
			})

			Context("when no pool is specified", func() {
				It("logs an error", func() {
					text := "@some-bot release"
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "no pool specified", text, channel)
				})
			})

			Context("when the pool is not claimed", func() {
				It("responds in slack", func() {
					text := fmt.Sprintf("@some-bot release %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}
					locker.StatusReturns([]string{}, []string{}, nil)

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())

					Expect(locker.ReleaseLockCallCount()).To(Equal(0))

					postChannel, postMessage := slackClient.PostMessageArgsForCall(0)
					Expect(slackClient.PostMessageCallCount()).To(Equal(1))
					Expect(postChannel).To(Equal(channel))
					Expect(postMessage).To(Equal(pool + " is not claimed"))
				})
			})

			Context("when checking the status fails", func() {
				It("logs an error", func() {
					text := "@some-bot release some-pool"
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}
					locker.StatusReturns(nil, nil, errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel)
				})
			})

			Context("when posting error to slack fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot release %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}
					locker.StatusReturns([]string{}, []string{}, nil)
					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel)
				})
			})

			Context("when releasing the lock fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot release %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}
					locker.StatusReturns([]string{pool}, []string{}, nil)

					locker.ReleaseLockReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel)
				})
			})

			Context("when posting to slack fails", func() {
				It("logs an error", func() {
					text := fmt.Sprintf("@some-bot release %s", pool)
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}
					locker.StatusReturns([]string{pool}, []string{}, nil)

					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel)
				})
			})
		})

		Context("when a status command is received", func() {
			It("posts the status of locks to slack", func() {
				text := "@some-bot status"
				slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
					messageHandler(text, channel)
					return nil
				}

				locker.StatusReturns(
					[]string{"claimed-1", "claimed-2"},
					[]string{"unclaimed-1", "unclaimed-2"},
					nil,
				)

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())

				Expect(locker.StatusCallCount()).To(Equal(1))

				postChannel, postMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(postChannel).To(Equal(channel))
				Expect(postMessage).To(Equal("*Claimed:* claimed-1, claimed-2\n*Unclaimed:* unclaimed-1, unclaimed-2"))
			})

			Context("when getting the status fails", func() {
				It("logs an error", func() {
					text := "@some-bot status"
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}

					locker.StatusReturns(nil, nil, errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel)
				})
			})

			Context("when posting to slack fails", func() {
				It("logs an error", func() {
					text := "@some-bot status"
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}

					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel)
				})
			})
		})

		Context("when a help command is received", func() {
			It("posts the help message to slack", func() {
				text := "@some-bot help"
				slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
					messageHandler(text, channel)
					return nil
				}

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())

				postChannel, postMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(postChannel).To(Equal(channel))
				Expect(postMessage).To(Equal(
					"Available commands:\n" +
						"```\n" +
						"  claim <env>     Claim an unclaimed environment\n" +
						"  release <env>   Release a claimed environment\n" +
						"  status          Show claimed and unclaimed environments\n" +
						"  help            Display this message\n" +
						"```",
				))
			})

			Context("when posting to slack fails", func() {
				It("logs an error", func() {
					text := "@some-bot help"
					slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
						messageHandler(text, channel)
						return nil
					}

					slackClient.PostMessageReturns(errors.New("some-error"))

					Expect(New(locker, slackClient, logger).Run()).To(Succeed())
					expectLoggedMessageHandlingError(logHook, "some-error", text, channel)
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
			It("logs an error", func() {
				text := "@some-bot"
				slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
					messageHandler(text, channel)
					return nil
				}

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())
				expectLoggedMessageHandlingError(logHook, "no command specified", text, channel)
			})
		})

		Context("when an unknown command is received", func() {
			It("logs an error", func() {
				text := "@some-bot some-bad-command"
				slackClient.ListenStub = func(messageHandler func(_, _ string)) error {
					messageHandler(text, channel)
					return nil
				}

				Expect(New(locker, slackClient, logger).Run()).To(Succeed())
				expectLoggedMessageHandlingError(logHook, "unknown command 'some-bad-command'", text, channel)
			})
		})
	})
})

func expectLoggedMessageHandlingError(logHook *logrustest.Hook, err string, text string, channel string) {
	ExpectWithOffset(1, len(logHook.Entries)).To(Equal(1))
	ExpectWithOffset(1, logHook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
	ExpectWithOffset(1, logHook.LastEntry().Message).To(Equal("Failed to handle message"))
	ExpectWithOffset(1, logHook.LastEntry().Data["error"]).To(Equal(err))
	ExpectWithOffset(1, logHook.LastEntry().Data["text"]).To(Equal(text))
	ExpectWithOffset(1, logHook.LastEntry().Data["channel"]).To(Equal(channel))
}

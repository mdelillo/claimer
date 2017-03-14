package claimer_test

import (
	. "github.com/mdelillo/claimer/claimer"

	"errors"
	"fmt"
	"github.com/mdelillo/claimer/claimer/claimerfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Claimer", func() {
	Describe("Run", func() {
		var (
			locker      *claimerfakes.FakeLocker
			slackClient *claimerfakes.FakeSlackClient
		)

		BeforeEach(func() {
			locker = new(claimerfakes.FakeLocker)
			slackClient = new(claimerfakes.FakeSlackClient)
		})

		Context("when a claim message is received", func() {
			It("claims the lock and posts to slack", func() {
				channel := "some-channel"
				pool := "some-pool"

				slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
					messageHandler(fmt.Sprintf("@some-bot claim %s", pool), channel)
					return nil
				}
				locker.StatusReturns([]string{}, []string{pool}, nil)

				claimer := New(locker, slackClient)
				Expect(claimer.Run()).To(Succeed())

				Expect(locker.ClaimLockCallCount()).To(Equal(1))
				Expect(locker.ClaimLockArgsForCall(0)).To(Equal(pool))

				postChannel, postMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(postChannel).To(Equal(channel))
				Expect(postMessage).To(Equal("Claimed " + pool))
			})

			Context("when no pool is specified", func() {
				It("returns an error", func() {
					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler("@some-bot claim", "some-channel")
						return nil
					}

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(MatchError("no pool specified"))
				})
			})

			Context("when the pool is not available", func() {
				It("responds in slack", func() {
					channel := "some-channel"
					pool := "some-pool"

					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler(fmt.Sprintf("@some-bot claim %s", pool), channel)
						return nil
					}
					locker.StatusReturns([]string{}, []string{}, nil)

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(Succeed())

					Expect(locker.ClaimLockCallCount()).To(Equal(0))

					postChannel, postMessage := slackClient.PostMessageArgsForCall(0)
					Expect(slackClient.PostMessageCallCount()).To(Equal(1))
					Expect(postChannel).To(Equal(channel))
					Expect(postMessage).To(Equal(pool + " is not available"))
				})
			})

			Context("when checking the status fails", func() {
				It("returns an error", func() {
					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler("@some-bot claim some-pool", "some-channel")
						return nil
					}
					locker.StatusReturns(nil, nil, errors.New("some-error"))

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(MatchError("some-error"))
				})
			})

			Context("when posting unavailability to slack fails", func() {
				It("returns an error", func() {
					pool := "some-pool"

					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler(fmt.Sprintf("@some-bot claim %s", pool), "some-channel")
						return nil
					}
					locker.StatusReturns([]string{}, []string{pool}, nil)
					slackClient.PostMessageReturns(errors.New("some-error"))

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(MatchError("some-error"))
				})
			})

			Context("when claiming the lock fails", func() {
				It("returns an error", func() {
					pool := "some-pool"

					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler(fmt.Sprintf("@some-bot claim %s", pool), "some-channel")
						return nil
					}
					locker.StatusReturns([]string{}, []string{pool}, nil)
					locker.ClaimLockReturns(errors.New("some-error"))

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(MatchError("some-error"))
				})
			})

			Context("when posting success to slack fails", func() {
				It("returns an error", func() {
					pool := "some-pool"

					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler(fmt.Sprintf("@some-bot claim %s", pool), "some-channel")
						return nil
					}
					locker.StatusReturns([]string{}, []string{pool}, nil)
					slackClient.PostMessageReturns(errors.New("some-error"))

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(MatchError("some-error"))
				})
			})
		})

		Context("when a release message is received", func() {
			It("releases the lock and posts to slack", func() {
				channel := "some-channel"
				pool := "some-pool"

				slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
					messageHandler(fmt.Sprintf("@some-bot release %s", pool), channel)
					return nil
				}

				claimer := New(locker, slackClient)
				Expect(claimer.Run()).To(Succeed())

				Expect(locker.ReleaseLockCallCount()).To(Equal(1))
				Expect(locker.ReleaseLockArgsForCall(0)).To(Equal(pool))

				postChannel, postMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(postChannel).To(Equal(channel))
				Expect(postMessage).To(Equal("Released " + pool))
			})

			Context("when no pool is specified", func() {
				It("returns an error", func() {
					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler("@some-bot release", "some-channel")
						return nil
					}

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(MatchError("no pool specified"))
				})
			})

			Context("when releasing the lock fails", func() {
				It("returns an error", func() {
					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler("@some-bot release some-pool", "some-channel")
						return nil
					}

					locker.ReleaseLockReturns(errors.New("some-error"))

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(MatchError("some-error"))
				})
			})

			Context("when posting to slack fails", func() {
				It("returns an error", func() {
					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler("@some-bot release some-pool", "some-channel")
						return nil
					}

					slackClient.PostMessageReturns(errors.New("some-error"))

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(MatchError("some-error"))
				})
			})
		})

		Context("when a status message is received", func() {
			It("posts the status of locks to slack", func() {
				channel := "some-channel"

				slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
					messageHandler("@some-bot status", channel)
					return nil
				}

				locker.StatusReturns(
					[]string{"claimed-1", "claimed-2"},
					[]string{"unclaimed-1", "unclaimed-2"},
					nil,
				)

				claimer := New(locker, slackClient)
				Expect(claimer.Run()).To(Succeed())

				Expect(locker.StatusCallCount()).To(Equal(1))

				postChannel, postMessage := slackClient.PostMessageArgsForCall(0)
				Expect(slackClient.PostMessageCallCount()).To(Equal(1))
				Expect(postChannel).To(Equal(channel))
				Expect(postMessage).To(Equal("*Claimed:* claimed-1, claimed-2\n*Unclaimed:* unclaimed-1, unclaimed-2"))
			})

			Context("when getting the status fails", func() {
				It("returns an error", func() {
					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler("@some-bot status", "some-channel")
						return nil
					}

					locker.StatusReturns(nil, nil, errors.New("some-error"))

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(MatchError("some-error"))
				})
			})

			Context("when posting to slack fails", func() {
				It("returns an error", func() {
					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler("@some-bot status", "some-channel")
						return nil
					}

					slackClient.PostMessageReturns(errors.New("some-error"))

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(MatchError("some-error"))
				})
			})
		})

		Context("when listening to slack messages fails", func() {
			It("returns an error", func() {
				slackClient.ListenReturns(errors.New("some-error"))
				claimer := New(locker, slackClient)
				Expect(claimer.Run()).To(MatchError("some-error"))
			})
		})

		Context("when no command is specified", func() {
			It("returns an error", func() {
				slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
					messageHandler("@some-bot", "some-channel")
					return nil
				}

				claimer := New(locker, slackClient)
				Expect(claimer.Run()).To(MatchError("no command specified"))
			})
		})

		Context("when an unknown command is received", func() {
			It("returns an error", func() {
				slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
					messageHandler("@some-bot some-bad-command", "some-channel")
					return nil
				}

				claimer := New(locker, slackClient)
				Expect(claimer.Run()).To(MatchError("unknown command 'some-bad-command'"))
			})
		})
	})
})

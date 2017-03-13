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

			Context("when claiming the lock fails", func() {
				It("returns an error", func() {
					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler("@some-bot claim some-pool", "some-channel")
						return nil
					}

					locker.ClaimLockReturns(errors.New("some-error"))

					claimer := New(locker, slackClient)
					Expect(claimer.Run()).To(MatchError("some-error"))
				})
			})

			Context("when posting to slack fails", func() {
				It("returns an error", func() {
					slackClient.ListenStub = func(messageHandler func(text, channel string)) error {
						messageHandler("@some-bot claim some-pool", "some-channel")
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

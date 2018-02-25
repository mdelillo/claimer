package commands_test

import (
	. "github.com/mdelillo/claimer/bot/commands"

	"errors"
	"fmt"

	"github.com/mdelillo/claimer/bot/commands/commandsfakes"
	clocker "github.com/mdelillo/claimer/locker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OwnerCommand", func() {
	Describe("Execute", func() {
		var locker *commandsfakes.FakeLocker

		BeforeEach(func() {
			locker = new(commandsfakes.FakeLocker)
		})

		Context("when the lock is claimed with a message", func() {
			It("responds with the owner, date, and message of the lock", func() {
				pool := "some-pool"
				owner := "some-owner"
				claimDate := "some-date"
				message := "some message"

				locker.StatusReturns(
					[]clocker.Lock{
						{Name: pool, Claimed: true, Owner: owner, Date: claimDate, Message: message},
					},
					nil,
				)

				command := NewFactory(locker).NewCommand("owner", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(fmt.Sprintf("%s was claimed by %s on %s (%s)", pool, owner, claimDate, message)))
			})
		})

		Context("when the lock is claimed without a message", func() {
			It("responds with the owner and date of the lock", func() {
				pool := "some-pool"
				owner := "some-owner"
				claimDate := "some-date"

				locker.StatusReturns(
					[]clocker.Lock{
						{Name: pool, Claimed: true, Owner: owner, Date: claimDate},
					},
					nil,
				)

				command := NewFactory(locker).NewCommand("owner", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(fmt.Sprintf("%s was claimed by %s on %s", pool, owner, claimDate)))
			})
		})

		Context("when the pool does not exist", func() {
			It("returns a slack response", func() {
				pool := "some-pool"

				locker.StatusReturns(nil, nil)

				command := NewFactory(locker).NewCommand("owner", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(pool + " does not exist"))
			})
		})

		Context("when the lock is not claimed", func() {
			It("returns a slack response", func() {
				pool := "some-pool"

				locker.StatusReturns(
					[]clocker.Lock{{Name: pool, Claimed: false}},
					nil,
				)

				command := NewFactory(locker).NewCommand("owner", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(pool + " is not claimed"))
			})
		})

		Context("when no pool is specified", func() {
			It("returns a slack response", func() {
				command := NewFactory(locker).NewCommand("owner", "", "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal("must specify pool"))
			})
		})

		Context("when checking the status fails", func() {
			It("returns an error", func() {
				pool := "some-pool"

				locker.StatusReturns(nil, errors.New("some-error"))

				command := NewFactory(locker).NewCommand("owner", pool, "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("failed to get status of locks: some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})
	})
})

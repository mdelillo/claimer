package commands_test

import (
	. "github.com/mdelillo/claimer/bot/commands"

	"errors"
	"fmt"
	"github.com/mdelillo/claimer/bot/commands/commandsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OwnerCommand", func() {
	Describe("Execute", func() {
		var locker *commandsfakes.FakeLocker

		BeforeEach(func() {
			locker = new(commandsfakes.FakeLocker)
		})

		Context("when the lock is claimed", func() {
			It("responds with the owner of the lock", func() {
				pool := "some-pool"
				owner := "some-owner"
				claimDate := "some-date"

				locker.StatusReturns([]string{pool}, []string{}, nil)
				locker.OwnerReturns(owner, claimDate, nil)

				command := NewFactory(locker).NewCommand("owner", []string{pool}, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(fmt.Sprintf("%s was claimed by %s on %s", pool, owner, claimDate)))

				Expect(locker.OwnerCallCount()).To(Equal(1))
				Expect(locker.OwnerArgsForCall(0)).To(Equal(pool))
			})
		})

		Context("when the lock is not claimed", func() {
			It("returns a slack response", func() {
				pool := "some-pool"

				locker.StatusReturns([]string{}, []string{}, nil)

				command := NewFactory(locker).NewCommand("owner", []string{pool}, "")

				slackResponse, err := command.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(slackResponse).To(Equal(pool + " is not claimed"))
			})
		})

		Context("when no pool is specified", func() {
			It("returns an error", func() {
				command := NewFactory(locker).NewCommand("owner", []string{}, "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("no pool specified"))
				Expect(slackResponse).To(BeEmpty())
			})
		})

		Context("when checking the status fails", func() {
			It("returns an error", func() {
				pool := "some-pool"

				locker.StatusReturns([]string{pool}, []string{}, errors.New("some-error"))

				command := NewFactory(locker).NewCommand("owner", []string{pool}, "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})

		Context("when checking the owner fails", func() {
			It("returns an error", func() {
				pool := "some-pool"

				locker.StatusReturns([]string{pool}, []string{}, nil)
				locker.OwnerReturns("", "", errors.New("some-error"))

				command := NewFactory(locker).NewCommand("owner", []string{pool}, "")

				slackResponse, err := command.Execute()
				Expect(err).To(MatchError("some-error"))
				Expect(slackResponse).To(BeEmpty())
			})
		})
	})
})

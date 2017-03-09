package slack_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSlack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Slack Suite")
}

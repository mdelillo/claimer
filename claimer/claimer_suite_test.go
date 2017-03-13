package claimer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestClaimer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Claimer Suite")
}

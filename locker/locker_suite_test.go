package locker_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLocker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Locker Suite")
}

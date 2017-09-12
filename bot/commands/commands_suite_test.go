package commands_test

import (
	"github.com/mdelillo/claimer/translate"
	"github.com/mdelillo/claimer/translations"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCommands(t *testing.T) {
	BeforeSuite(func() {
		Expect(translate.LoadTranslations(translations.DefaultTranslations)).To(Succeed())
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}

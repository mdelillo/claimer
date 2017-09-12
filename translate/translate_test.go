package translate_test

import (
	. "github.com/mdelillo/claimer/translate"

	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Translate", func() {
	Describe("Translation", func() {
		var translationsPath string

		AfterEach(func() {
			os.RemoveAll(translationsPath)
		})

		Context("load from a string", func() {
			It("translates using the provided yaml string", func() {
				translation := "key1: val1"

				Expect(LoadTranslations(translation)).To(Succeed())
				Expect(T("key1", nil)).To(Equal("val1"))
			})

		})

		Context("multiple translation files", func() {
			var otherTranslationsPath string

			BeforeEach(func() {
				translationsPath = writeTranslationFile(`---
key1: val1
parent:
  child1: child-val1
  child2: child-val2`)
				otherTranslationsPath = writeTranslationFile(`---
key2: val2
parent:
  child2: other-child-val2
  child3: other-child-val3`)
			})

			AfterEach(func() {
				os.RemoveAll(otherTranslationsPath)
			})

			It("translates using only the last loaded file", func() {
				Expect(LoadTranslationFile(translationsPath)).To(Succeed())
				Expect(LoadTranslationFile(otherTranslationsPath)).To(Succeed())
				Expect(T("key1", nil)).To(Equal("key1"))
				Expect(T("key2", nil)).To(Equal("val2"))
				Expect(T("parent.child1", nil)).To(Equal("parent.child1"))
				Expect(T("parent.child2", nil)).To(Equal("other-child-val2"))
				Expect(T("parent.child3", nil)).To(Equal("other-child-val3"))
			})
		})

		Context("passing variables to translation", func() {
			BeforeEach(func() {
				translationsPath = writeTranslationFile("key: some {{.var1}} interpolated {{.var2}} string")
				Expect(LoadTranslationFile(translationsPath)).To(Succeed())
			})

			It("interpolates the variables into the translated string", func() {
				vars := TArgs{
					"var1": "value1",
					"var2": "value2",
				}
				Expect(T("key", vars)).To(Equal("some value1 interpolated value2 string"))
			})

			It("does not interpolate variables that are not passed in", func() {
				vars := TArgs{
					"var2": "value2",
				}
				Expect(T("key", vars)).To(Equal("some {{.var1}} interpolated value2 string"))
			})
		})

		Context("translation file errors", func() {
			Context("when the translation file does not exist", func() {
				It("returns an error", func() {
					Expect(LoadTranslationFile("some-bad-path")).To(MatchError("failed to read file: some-bad-path"))
				})
			})

			Context("when the translation file is not valid YAML", func() {
				BeforeEach(func() {
					translationsPath = writeTranslationFile("some-invalid-yaml")
				})

				It("returns an error", func() {
					Expect(LoadTranslationFile(translationsPath)).To(MatchError("failed to parse YAML: some-invalid-yaml"))
				})
			})
		})

		Context("key errors", func() {
			Context("when the key does not exit", func() {
				BeforeEach(func() {
					translationsPath = writeTranslationFile("---")
				})

				It("returns the untranslated key", func() {
					Expect(LoadTranslationFile(translationsPath)).To(Succeed())
					Expect(T("missingkey", nil)).To(Equal("missingkey"))
				})
			})

			Context("when the nested key does not exit", func() {
				BeforeEach(func() {
					translationsPath = writeTranslationFile("nested: {key: some-value}")
				})

				It("returns the untranslated key", func() {
					Expect(LoadTranslationFile(translationsPath)).To(Succeed())
					Expect(T("nested.missingkey", nil)).To(Equal("nested.missingkey"))
				})
			})

			Context("when a value is a string instead of a nested map", func() {
				BeforeEach(func() {
					translationsPath = writeTranslationFile("nested: {key: value}")
				})

				It("returns the untranslated key", func() {
					Expect(LoadTranslationFile(translationsPath)).To(Succeed())
					Expect(T("nested.key.otherkey", nil)).To(Equal("nested.key.otherkey"))
				})
			})

			Context("when a value is not a string or map", func() {
				BeforeEach(func() {
					translationsPath = writeTranslationFile("key: [not-a-string-or-map]")
				})

				It("returns the untranslated key", func() {
					Expect(LoadTranslationFile(translationsPath)).To(Succeed())
					Expect(T("key", nil)).To(Equal("key"))
				})
			})
		})
	})
})

func writeTranslationFile(translations string) string {
	file, err := ioutil.TempFile("", "claimer-translate")
	Expect(err).NotTo(HaveOccurred())
	Expect(ioutil.WriteFile(file.Name(), []byte(translations), 0644)).To(Succeed())
	return file.Name()
}

package requests_test

import (
	. "github.com/mdelillo/claimer/slack/requests"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http/httptest"
	"fmt"
	"net/http"
)

var _ = Describe("UsernameRequest", func() {
	Describe("Execute", func() {
		It("returns username for the given user id", func() {
			apiToken := "some-api-token"
			userId := "some-user-id"
			username := "some-username"

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()

				Expect(r.RequestURI).To(Equal(fmt.Sprintf("/api/users.info?token=%s&user=%s", apiToken, userId)))
				Expect(r.Method).To(Equal("GET"))

				w.Write([]byte(fmt.Sprintf(`{"ok": true, "user": {"name": "%s"}}`, username)))
			}))
			defer server.Close()

			request := NewFactory(server.URL, apiToken).NewUsernameRequest(userId)
			actualUsername, err := request.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualUsername).To(Equal(username))
		})
	})

	Context("when the request fails", func() {
		It("returns an error", func() {
			_, err := NewFactory("", "").NewUsernameRequest("").Execute()
			Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
		})
	})

	Context("when the status code is not 200", func() {
		It("returns an error", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()
				w.WriteHeader(503)
			}))
			defer server.Close()

			_, err := NewFactory(server.URL, "").NewUsernameRequest("").Execute()
			Expect(err).To(MatchError(ContainSubstring("error getting user info: 503 Service Unavailable")))
		})
	})

	Context("when unmarshaling the body fails", func() {
		It("returns an error", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()

				w.Write([]byte(`some-bad-json`))
			}))
			defer server.Close()

			_, err := NewFactory(server.URL, "").NewUsernameRequest("").Execute()
			Expect(err).To(MatchError(ContainSubstring("invalid character")))
		})
	})

	Context("when the response is an error", func() {
		It("returns an error", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()

				w.Write([]byte(`{"ok": false, "error": "some-error"}`))
			}))
			defer server.Close()

			_, err := NewFactory(server.URL, "").NewUsernameRequest("").Execute()
			Expect(err).To(MatchError("error in slack response: some-error"))
		})
	})
})

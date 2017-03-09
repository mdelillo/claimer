package slack_test

import (
	"github.com/mdelillo/claimer/slack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Client", func() {
	Describe("PostMessage", func() {
		It("posts a message to a slack channel", func() {
			apiToken := "some-api-token"
			channel := "some-channel"
			message := "some-message"

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()

				Expect(r.RequestURI).To(Equal("/api/chat.postMessage"))
				Expect(r.Method).To(Equal("POST"))
				Expect(r.FormValue("token")).To(Equal(apiToken))
				Expect(r.FormValue("channel")).To(Equal(channel))
				Expect(r.FormValue("text")).To(Equal(message))
				Expect(r.FormValue("as_user")).To(Equal("true"))

				w.Write([]byte(`{"ok": true}`))
			}))
			defer server.Close()

			client := slack.NewClient(server.URL, apiToken)
			Expect(client.PostMessage(channel, message)).To(Succeed())
		})

		Context("when the response code is non-200", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					w.WriteHeader(503)
				}))
				defer server.Close()

				client := slack.NewClient(server.URL, "")
				err := client.PostMessage("", "")
				Expect(err).To(MatchError("error posting to slack: 503 Service Unavailable"))
			})
		})

		Context("when slack returns an error", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					w.Write([]byte(`{"ok": false, "error": "some-error"}`))
				}))
				defer server.Close()

				client := slack.NewClient(server.URL, "")
				err := client.PostMessage("", "")
				Expect(err).To(MatchError("error posting to slack: some-error"))
			})
		})

		Context("when the response from slack cannot be parsed", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					w.Write([]byte("some-invalid-json"))
				}))
				defer server.Close()

				client := slack.NewClient(server.URL, "")
				err := client.PostMessage("", "")
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})
		})
	})
})

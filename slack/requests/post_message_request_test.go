package requests_test

import (
	. "github.com/mdelillo/claimer/slack/requests"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("PostMessageRequest", func() {
	It("posts a message to slack", func() {
		apiToken := "some-api-token"
		channel := "some-channel"
		message := "some-message"
		messageReceived := false

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()

			Expect(r.RequestURI).To(Equal("/api/chat.postMessage"))
			Expect(r.Method).To(Equal("POST"))
			Expect(r.FormValue("token")).To(Equal(apiToken))
			Expect(r.FormValue("channel")).To(Equal(channel))
			Expect(r.FormValue("text")).To(Equal(message))
			Expect(r.FormValue("as_user")).To(Equal("true"))

			w.Write([]byte(`{"ok": true}`))
			messageReceived = true
		}))
		defer server.Close()

		request := NewFactory(server.URL, apiToken).NewPostMessageRequest(channel, message)
		Expect(request.Execute()).To(Succeed())
		Expect(messageReceived).To(BeTrue())
	})

	Context("when the request fails", func() {
		It("returns an error", func() {
			err := NewFactory("", "").NewPostMessageRequest("", "").Execute()
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

			err := NewFactory(server.URL, "").NewPostMessageRequest("", "").Execute()
			Expect(err).To(MatchError("error posting message: 503 Service Unavailable"))
		})
	})

	Context("when unmarshaling the body fails", func() {
		It("returns an error", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()
				w.Write([]byte("some-bad-json"))
			}))
			defer server.Close()

			err := NewFactory(server.URL, "").NewPostMessageRequest("", "").Execute()
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

			err := NewFactory(server.URL, "").NewPostMessageRequest("", "").Execute()
			Expect(err).To(MatchError("error in slack response: some-error"))
		})
	})
})

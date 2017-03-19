package requests_test

import (
	. "github.com/mdelillo/claimer/slack/requests"

	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("StartRtmRequest", func() {
	Describe("Execute", func() {
		It("starts an RTM session and returns a websocket URL and bot ID", func() {
			apiToken := "some-api-token"
			botId := "some-bot-id"
			websocketUrl := "some-websocket-url"

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()

				Expect(r.RequestURI).To(Equal("/api/rtm.start?token=" + apiToken))
				Expect(r.Method).To(Equal("GET"))

				w.Write([]byte(fmt.Sprintf(
					`{"ok": true, "url": "%s", "self": {"id": "%s"}}`,
					websocketUrl,
					botId,
				)))
			}))
			defer server.Close()

			request := NewFactory(server.URL, apiToken).NewStartRtmRequest()
			actualWebsocketUrl, actualBotId, err := request.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualWebsocketUrl).To(Equal(websocketUrl))
			Expect(actualBotId).To(Equal(botId))
		})

		Context("when the request fails", func() {
			It("returns an error", func() {
				_, _, err := NewFactory("", "").NewStartRtmRequest().Execute()
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

				_, _, err := NewFactory(server.URL, "").NewStartRtmRequest().Execute()
				Expect(err).To(MatchError("error starting RTM session: 503 Service Unavailable"))
			})
		})

		Context("when unmarshaling the body fails", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()
					w.Write([]byte("some-bad-json"))
				}))
				defer server.Close()

				_, _, err := NewFactory(server.URL, "").NewStartRtmRequest().Execute()
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

				_, _, err := NewFactory(server.URL, "").NewStartRtmRequest().Execute()
				Expect(err).To(MatchError("error in slack response: some-error"))
			})
		})
	})
})

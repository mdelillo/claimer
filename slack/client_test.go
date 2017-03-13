package slack_test

import (
	. "github.com/mdelillo/claimer/slack"

	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/websocket"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Client", func() {
	Describe("Listen", func() {
		It("handles incoming messages mentioning the user using the supplied function", func() {
			apiToken := "some-api-token"
			botId := "some-bot-id"
			channel := "some-channel"

			websocketServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
				ws.Write([]byte(fmt.Sprintf(
					`{"type":"%s", "text":"%s", "channel":"%s"}`,
					"message",
					fmt.Sprintf("<@%s> some-text", botId),
					channel,
				)))
				ws.Write([]byte(fmt.Sprintf(
					`{"type":"%s", "text":"%s", "channel":"%s"}`,
					"not-a-message",
					"some-non-message-text",
					channel,
				)))
				ws.Write([]byte(fmt.Sprintf(
					`{"type":"%s", "text":"%s", "channel":"%s"}`,
					"message",
					"<@some-other-id> some-text-for-other-user",
					channel,
				)))
				ws.Write([]byte(fmt.Sprintf(
					`{"type":"%s", "text":"%s", "channel":"%s"}`,
					"message",
					fmt.Sprintf("<@%s> some-other-text", botId),
					channel,
				)))
			}))
			defer websocketServer.Close()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()

				Expect(r.RequestURI).To(Equal("/api/rtm.start?token=" + apiToken))
				Expect(r.Method).To(Equal("GET"))

				w.Write([]byte(fmt.Sprintf(
					`{"ok": true, "url": "%s", "self": {"id": "%s"}}`,
					"ws://"+websocketServer.Listener.Addr().String(),
					botId,
				)))
			}))
			defer server.Close()

			messageCount := 0
			messageHandler := func(text, channel string) {
				messageCount++
				Expect(text).To(HavePrefix(fmt.Sprintf("<@%s>", botId)))
				Expect(channel).To(Equal(channel))
			}

			NewClient(server.URL, apiToken).Listen(messageHandler)
			Eventually(func() int { return messageCount }).Should(Equal(2))
			Eventually(func() int { return messageCount }).ShouldNot(Equal(3))
		})

		Context("when there is an error starting the RTM session", func() {
			Context("when the response code is non-200", func() {
				It("returns an error", func() {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						defer GinkgoRecover()

						w.WriteHeader(503)
					}))
					defer server.Close()

					client := NewClient(server.URL, "")
					Expect(client.Listen(nil)).To(MatchError("bad response code: 503 Service Unavailable"))
				})
			})

			Context("when slack returns an error", func() {
				It("returns an error", func() {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						defer GinkgoRecover()

						w.Write([]byte(`{"ok": false, "error": "some-error"}`))
					}))
					defer server.Close()

					client := NewClient(server.URL, "")
					Expect(client.Listen(nil)).To(MatchError("failed to start RTM session: some-error"))
				})
			})

			Context("when the response from slack cannot be parsed", func() {
				It("returns an error", func() {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						defer GinkgoRecover()

						w.Write([]byte("some-invalid-json"))
					}))
					defer server.Close()

					client := NewClient(server.URL, "")
					Expect(client.Listen(nil)).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})

		Context("when connecting to the websocket fails", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					w.Write([]byte(`{"ok": true, "url": "some-bad-url", "self": {"id": "some-bot-id"}}`))
				}))
				defer server.Close()

				client := NewClient(server.URL, "")
				Expect(client.Listen(nil)).To(MatchError(MatchRegexp("failed to connect to websocket: .*some-bad-url.*")))
			})
		})

		Context("when parsing the message fails", func() {
			It("returns an error", func() {
				websocketServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
					ws.Write([]byte("some-bad-data"))
				}))
				defer websocketServer.Close()

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					w.Write([]byte(fmt.Sprintf(
						`{"ok": true, "url": "%s", "self": {"id": "some-bot-id"}}`,
						"ws://"+websocketServer.Listener.Addr().String(),
					)))
				}))
				defer server.Close()

				client := NewClient(server.URL, "")
				Expect(client.Listen(nil)).To(MatchError(ContainSubstring("failed to parse message: ")))
			})
		})
	})

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

			client := NewClient(server.URL, apiToken)
			Expect(client.PostMessage(channel, message)).To(Succeed())
		})

		Context("when the response code is non-200", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					w.WriteHeader(503)
				}))
				defer server.Close()

				client := NewClient(server.URL, "")
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

				client := NewClient(server.URL, "")
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

				client := NewClient(server.URL, "")
				err := client.PostMessage("", "")
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})
		})
	})
})

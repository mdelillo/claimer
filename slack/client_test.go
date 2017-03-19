package slack_test

import (
	. "github.com/mdelillo/claimer/slack"

	"errors"
	"fmt"
	"github.com/mdelillo/claimer/slack/requests/requestsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/websocket"
	"net/http/httptest"
)

var _ = Describe("Client", func() {
	var (
		requestFactory     *requestsfakes.FakeFactory
		getUsernameRequest *requestsfakes.FakeGetUsernameRequest
		postMessageRequest *requestsfakes.FakePostMessageRequest
		startRtmRequest    *requestsfakes.FakeStartRtmRequest
	)

	BeforeEach(func() {
		requestFactory = new(requestsfakes.FakeFactory)
		getUsernameRequest = new(requestsfakes.FakeGetUsernameRequest)
		postMessageRequest = new(requestsfakes.FakePostMessageRequest)
		startRtmRequest = new(requestsfakes.FakeStartRtmRequest)
	})

	Describe("Listen", func() {
		It("handles incoming messages mentioning the user using the supplied function", func() {
			botId := "some-bot-id"
			channel := "some-channel"
			userId := "some-user-id"
			username := "some-username"

			websocketServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
				ws.Write([]byte(fmt.Sprintf(
					`{"type":"%s", "text":"%s", "channel":"%s", "user":"%s"}`,
					"message",
					fmt.Sprintf("<@%s> some-text", botId),
					channel,
					userId,
				)))
				ws.Write([]byte(fmt.Sprintf(
					`{"type":"%s", "text":"%s", "channel":"%s", "user":"%s"}`,
					"not-a-message",
					"some-non-message-text",
					channel,
					userId,
				)))
				ws.Write([]byte(fmt.Sprintf(
					`{"type":"%s", "text":"%s", "channel":"%s", "user":"%s"}`,
					"message",
					"<@some-other-id> some-text-for-other-user",
					channel,
					userId,
				)))
				ws.Write([]byte(fmt.Sprintf(
					`{"type":"%s", "text":"%s", "channel":"%s", "user":"%s"}`,
					"message",
					fmt.Sprintf("<@%s> some-other-text", botId),
					channel,
					userId,
				)))
			}))
			defer websocketServer.Close()
			websocketUrl := "ws://" + websocketServer.Listener.Addr().String()

			requestFactory.NewStartRtmRequestReturns(startRtmRequest)
			startRtmRequest.ExecuteReturns(websocketUrl, botId, nil)

			requestFactory.NewGetUsernameRequestReturns(getUsernameRequest)
			getUsernameRequest.ExecuteReturns(username, nil)

			messageCount := 0
			messageHandler := func(actualText, actualChannel, actualUsername string) {
				messageCount++
				Expect(actualText).To(HavePrefix(fmt.Sprintf("<@%s>", botId)))
				Expect(actualChannel).To(Equal(channel))
				Expect(actualUsername).To(Equal(username))
			}

			NewClient(requestFactory).Listen(messageHandler)
			Eventually(func() int { return messageCount }).Should(Equal(2))
			Eventually(func() int { return messageCount }).ShouldNot(Equal(3))
			Expect(requestFactory.NewStartRtmRequestCallCount()).To(Equal(1))
			Expect(requestFactory.NewGetUsernameRequestCallCount()).To(Equal(2))
			Expect(requestFactory.NewGetUsernameRequestArgsForCall(0)).To(Equal(userId))
			Expect(requestFactory.NewGetUsernameRequestArgsForCall(1)).To(Equal(userId))
			Expect(getUsernameRequest.ExecuteCallCount()).To(Equal(2))
		})

		Context("when there is an error starting the RTM session", func() {
			It("returns an error", func() {
				requestFactory.NewStartRtmRequestReturns(startRtmRequest)
				startRtmRequest.ExecuteReturns("", "", errors.New("some-error"))

				client := NewClient(requestFactory)
				Expect(client.Listen(nil)).To(MatchError(MatchRegexp("some-error")))
			})
		})

		Context("when connecting to the websocket fails", func() {
			It("returns an error", func() {
				requestFactory.NewStartRtmRequestReturns(startRtmRequest)
				startRtmRequest.ExecuteReturns("some-bad-url", "some-bot-id", nil)

				client := NewClient(requestFactory)
				Expect(client.Listen(nil)).To(MatchError(MatchRegexp("failed to connect to websocket: .*some-bad-url.*")))
			})
		})

		Context("when parsing the event fails", func() {
			It("returns an error", func() {
				websocketServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
					ws.Write([]byte("some-bad-data"))
				}))
				defer websocketServer.Close()
				websocketUrl := "ws://" + websocketServer.Listener.Addr().String()

				requestFactory.NewStartRtmRequestReturns(startRtmRequest)
				startRtmRequest.ExecuteReturns(websocketUrl, "some-bot-id", nil)

				client := NewClient(requestFactory)
				Expect(client.Listen(nil)).To(MatchError(ContainSubstring("failed to parse event: ")))
			})
		})

		Context("when parsing the message fails", func() {
			It("returns an error", func() {
				websocketServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
					ws.Write([]byte(`{"type": "message", "text": {"bad-structure": true}}`))
				}))
				defer websocketServer.Close()
				websocketUrl := "ws://" + websocketServer.Listener.Addr().String()

				requestFactory.NewStartRtmRequestReturns(startRtmRequest)
				startRtmRequest.ExecuteReturns(websocketUrl, "some-bot-id", nil)

				client := NewClient(requestFactory)
				Expect(client.Listen(nil)).To(MatchError(ContainSubstring("failed to parse message: ")))
			})
		})

		Context("when getting the username fails", func() {
			It("returns an error", func() {
				botId := "some-bot-id"
				userId := "some-user-id"

				websocketServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
					ws.Write([]byte(fmt.Sprintf(
						`{"type":"%s", "text":"%s", "channel":"some-channel", "user":"%s"}`,
						"message",
						fmt.Sprintf("<@%s> some-text", botId),
						userId,
					)))
				}))
				defer websocketServer.Close()
				websocketUrl := "ws://" + websocketServer.Listener.Addr().String()

				requestFactory.NewStartRtmRequestReturns(startRtmRequest)
				startRtmRequest.ExecuteReturns(websocketUrl, botId, nil)

				requestFactory.NewGetUsernameRequestReturns(getUsernameRequest)
				getUsernameRequest.ExecuteReturns("", errors.New("some-error"))

				client := NewClient(requestFactory)
				Expect(client.Listen(nil)).To(MatchError(fmt.Sprintf("failed to get username for %s: some-error", userId)))
			})
		})
	})

	Describe("PostMessage", func() {
		It("makes a PostMessage request", func() {
			channel := "some-channel"
			message := "some-message"

			requestFactory.NewPostMessageRequestReturns(postMessageRequest)
			postMessageRequest.ExecuteReturns(nil)

			client := NewClient(requestFactory)
			Expect(client.PostMessage(channel, message)).To(Succeed())

			actualChannel, actualMessage := requestFactory.NewPostMessageRequestArgsForCall(0)
			Expect(actualChannel).To(Equal(channel))
			Expect(actualMessage).To(Equal(message))
		})

		Context("when the requets fails", func() {
			It("returns an error", func() {
				requestFactory.NewPostMessageRequestReturns(postMessageRequest)
				postMessageRequest.ExecuteReturns(errors.New("some-error"))

				client := NewClient(requestFactory)
				Expect(client.PostMessage("", "")).To(MatchError("some-error"))
			})
		})
	})
})

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/nlopes/slack"
)

const (
	// action is used for slack attachment action.
	actionSelect = "select"
	actionStart  = "start"
	actionCancel = "cancel"
)

type SlackListener struct {
	client    *slack.Client
	botID     string
	channelID string
}

// ListenAndResponse listens slack events and response
// particular messages. It replies by slack message button.
func (s *SlackListener) ListenAndResponse() {
	rtm := s.client.NewRTM()

	// Start listening slack events
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(ev); err != nil {
				log.Println("[ERROR] Failed to handle message: ", err)
			}
		}
	}
}

// handleMesageEvent handles message events.
func (s *SlackListener) handleMessageEvent(ev *slack.MessageEvent) error {
	// Only response in specific channel. Ignore else.
	if ev.Channel != s.channelID {
		log.Printf("%s != %s", ev.Channel, ev.Msg.Text)
		return nil
	}

	// Only response mention to bot. Ignore else.
	if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s> ", s.botID)) {
		return nil
	}

	// Parse message
	m := strings.Fields(strings.TrimSpace(ev.Msg.Text))[1:]
	if len(m) < 2 {
		return fmt.Errorf("invalid message")
	}

	GistUser = m[1]
	var attachment slack.Attachment
	switch m[0] {
	case "log":
		// value is passed to message handler when request is approved.
		attachment = slack.Attachment{
			Text:       ":octocat: Which period of Gist activities do you want to see? :date:",
			Color:      "#f9a41b",
			CallbackID: "log",
			Actions: []slack.AttachmentAction{
				{
					Name: actionSelect,
					Type: "select",
					Options: []slack.AttachmentActionOption{
						{
							Text:  "Weekly",
							Value: "weekly",
						},
						{
							Text:  "Monthly",
							Value: "monthly",
						},
						{
							Text:  "Yearly",
							Value: "yearly",
						},
					},
				},
				{
					Name:  actionCancel,
					Text:  "Cancel",
					Type:  "button",
					Style: "danger",
				},
			},
		}
	default:
		return fmt.Errorf("invalid message")
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			attachment,
		},
	}

	if _, _, err := s.client.PostMessage(ev.Channel, "", params); err != nil {
		return fmt.Errorf("failed to post message: %s", err)
	}

	return nil
}

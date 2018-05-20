package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/nlopes/slack"
)

// interactionHandler handles interactive message response.
type interactionHandler struct {
	slackClient       *slack.Client
	verificationToken string
	gistAccessToken   string
}

func (h interactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] Invalid method: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		log.Printf("[ERROR] Failed to unespace request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.AttachmentActionCallback
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		log.Printf("[ERROR] Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Only accept message from slack with valid token
	if message.Token != h.verificationToken {
		log.Printf("[ERROR] Invalid token: %s", message.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	action := message.Actions[0]
	switch action.Name {
	case actionSelect:
		Period = action.SelectedOptions[0].Value

		// Overwrite original drop down message.
		originalMessage := message.OriginalMessage
		originalMessage.Attachments[0].Text = fmt.Sprintf(":ledger: List %s activities?", strings.Title(Period))
		originalMessage.Attachments[0].Actions = []slack.AttachmentAction{
			{
				Name:  actionStart,
				Text:  "Yes",
				Type:  "button",
				Value: Period,
				Style: "primary",
			},
			{
				Name:  actionCancel,
				Text:  "No",
				Type:  "button",
				Style: "danger",
			},
		}

		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&originalMessage)
		return
	case actionStart:
		title := ":ok: retrieving Gist activities..."
		responseMessage(w, message.OriginalMessage, title, "")

		go responseGistHistory(h, message.ResponseURL, Period)

		return
	case actionCancel:
		title := fmt.Sprintf(":x: @%s canceled the request", message.User.Name)
		responseMessage(w, message.OriginalMessage, title, "")
		return
	default:
		log.Println("[ERROR] Invalid action was submitted: ", action.Name)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// responseMessage response to the original slackbutton enabled message.
// It removes button and replace it with message which indicate how bot will work
func responseMessage(w http.ResponseWriter, original slack.Message, title, value string) {
	original.Attachments[0].Actions = []slack.AttachmentAction{} // empty buttons
	original.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: title,
			Value: value,
			Short: false,
		},
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)
}

// delayMessage is used for messages sending after a short interval.
type delayMessage struct {
	Text  string `json:"text"`
	Token string `json:"token"`
}

// responseDelayMessage sends the delay message to the response URL.
func responseDelayMessage(h interactionHandler, endpoint, text string) error {
	message := delayMessage{Text: text, Token: h.verificationToken}
	reqBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to encode delayed message: %s", err)
	}

	res, err := http.Post(endpoint, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to send the delayed message: %s", err)
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to read the response against delayed message: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("[ERROR] Invalid response status: %d", res.StatusCode)
	}
	log.Println("[INFO] Response body: ", string(resBody))

	return nil
}

// responseDelayMessage sends the delay message to the response URL.
func responseGistHistory(h interactionHandler, endpoint, period string) error {
	text, err := getGistHistory(h, period)
	if err != nil {
		return err
	}

	if responseDelayMessage(h, endpoint, text) != nil {
		return err
	}

	return nil
}

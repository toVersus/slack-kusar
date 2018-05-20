package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
)

// https://api.slack.com/slack-apps
// https://api.slack.com/internal-integrations
type envConfig struct {
	// Port is server port to be listened.
	Port string `envconfig:"PORT" default:"3000"`

	// BotToken is bot user token to access to slack API.
	BotToken string `envconfig:"BOT_TOKEN" required:"true"`

	// VerificationToken is used to validate interactive messages from slack.
	VerificationToken string `envconfig:"VERIFICATION_TOKEN" required:"true"`

	// BotID is bot user ID.
	// You can find the channel ID from the following page:
	// https://api.slack.com/methods/users.list/test
	BotID string `envconfig:"BOT_ID" required:"true"`

	// ChannelID is slack channel ID where bot is working.
	// Bot responses to the mention in this channel.
	// You can find the channel ID from the following page:
	// https://api.slack.com/methods/users.list/test
	ChannelID string `envconfig:"CHANNEL_ID" required:"true"`

	// GistPersonalAccessToken is the tokens generated in the following page:
	// https://github.com/settings/tokens
	GistPersonalAccessToken string `envconfig:"GIST_ACCESS_TOKEN" required:"true"`
}

func main() {
	os.Exit(_main(os.Args[1:]))
}

func _main(args []string) int {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Println("[ERROR] Failed to parse env var: ", err)
		return 1
	}

	// Listening slack event and response
	log.Println("[INFO] Start slack event listening")
	slackListener := &SlackListener{
		client:    slack.New(env.BotToken),
		botID:     env.BotID,
		channelID: env.ChannelID,
	}
	go slackListener.ListenAndResponse()

	// Register handler to receive interactive message
	// responses from slack (kicked by user action)
	http.Handle("/interaction", interactionHandler{
		verificationToken: env.VerificationToken,
		gistAccessToken:   env.GistPersonalAccessToken,
	})

	log.Println("[INFO] Server listening on: ", env.Port)
	if err := http.ListenAndServe(":"+env.Port, nil); err != nil {
		log.Println("[ERROR] ", err)
		return 1
	}

	return 0
}

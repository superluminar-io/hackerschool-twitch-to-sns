package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/gempir/go-twitch-irc"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/jessevdk/go-flags"
)

type Msg struct {
	Channel   string    `json:"channel"`
	Username  string    `json:"username"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

func main() {

	var opts struct {
		TwitchUser     string   `short:"u" long:"twitch-user" required:"true" name:"twitch username"`
		TwitchOauth    string   `short:"o" long:"twitch-oauth" required:"true" name:"twitch oauth token"`
		TwitchChannels []string `short:"c" long:"twitch-channels" required:"true" name:"twitch channels to join"`
		SNSTopicArn    string   `short:"a" long:"sns-topic-arn" required:"true" name:"sns topic arn to publish messages to"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	aws_session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	sns_service := sns.New(aws_session)

	irc_client := twitch.NewClient(opts.TwitchUser, opts.TwitchOauth)
	irc_client.OnNewMessage(func(channel string, user twitch.User, message twitch.Message) {
		sns_msg := &Msg{
			Channel:   channel,
			Username:  user.DisplayName,
			Timestamp: message.Time,
			Message:   message.Text,
		}
		sns_msg_json, _ := json.Marshal(sns_msg)
		log.Println("incoming message", string(sns_msg_json))

		sns_params := &sns.PublishInput{
			Message:  aws.String(string(sns_msg_json)),
			TopicArn: aws.String(opts.SNSTopicArn),
		}
		sns_response, err := sns_service.Publish(sns_params)
		if err != nil {
			log.Println(err.Error())
		}

		log.Println("published message", sns_response)
	})

	for _, channel := range opts.TwitchChannels {
		log.Println("joining channel", channel)
		irc_client.Join(channel)
	}

	err = irc_client.Connect()
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}

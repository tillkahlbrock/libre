package main

import (
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/mailgun/mailgun-go"
)

// Your available domain names can be found here:
// (https://app.mailgun.com/app/domains)
var yourDomain string = "cerebral-kittens.com"

// You can find the Private API Key in your Account Menu, under "Settings":
// (https://app.mailgun.com/app/account/security)
var privateAPIKey string = "key-507z7kers44cwe1dm0g8fr3v8mwnpf05"

func Handler(event events.SNSEvent) error {
	// Create an instance of the Mailgun Client
	mg := mailgun.NewMailgun(yourDomain, privateAPIKey)
	c := Config{MGClient: mg}

	for _, r := range event.Records {
		if err := c.SendMail(r.SNS.Message); err != nil {
			logrus.WithError(err).Warnf("failed to send message")
		}
	}
	return nil
}

type Config struct {
	MGClient mailgun.Mailgun
}

func (c *Config) SendMail(body string) error {
	mg := c.MGClient
	sender := "libre@cerebral-kittens.com"
	subject := "Auf zur Bücherei!"
	recipient := "till.kahlbrock@gmail.com"

	message := mg.NewMessage(sender, subject, body, recipient)
	resp, id, err := mg.Send(message)
	if err != nil {
		return err
	}
	logrus.Infof("ID: %s Resp: %s", id, resp)
	return nil
}

func main() {
	lambda.Start(Handler)
}

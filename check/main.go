package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/sns/snsiface"

	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/aws/aws-lambda-go/lambda"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Item struct {
	bibNum  string
	title   string
	dueDate string
}

type Config struct {
	DynamoDBTable string
	DBClient      dynamodbiface.DynamoDBAPI
	SNSTopicArn   string
	SNSClient     snsiface.SNSAPI
}

type Sorted struct {
	Warn []Item
	Crit []Item
	Rest []Item
}

var AWSSession = session.Must(session.NewSession())

var config = Config{
	DynamoDBTable: os.Getenv("DYNAMODB_TABLE"),
	DBClient:      dynamodb.New(AWSSession),
	SNSTopicArn:   os.Getenv("SNS_TOPIC_ARN"),
	SNSClient:     sns.New(AWSSession),
}

func (c *Config) fetchItems() {
	output, err := c.DBClient.Scan(&dynamodb.ScanInput{TableName: aws.String(c.DynamoDBTable)})
	if err != nil {
		log.WithField("error", err).Fatalf("Failed to scan table '%s'", c.DynamoDBTable)
	}

	sorted := Sorted{
		Warn: []Item{},
		Crit: []Item{},
		Rest: []Item{},
	}
	for _, i := range output.Items {
		item := Item{
			bibNum:  *i["id"].S,
			title:   *i["title"].S,
			dueDate: *i["due_date"].S,
		}

		layout := "2006-01-02T15:04:05"
		parsed, err := time.Parse(layout, item.dueDate)
		if err != nil {
			fmt.Println(err)
		}

		until := time.Until(parsed)
		if until.Hours() < 24 {
			sorted.Crit = append(sorted.Crit, item)
		} else if until.Hours() < 72 {
			sorted.Warn = append(sorted.Warn, item)
		} else {
			sorted.Rest = append(sorted.Rest, item)
		}

	}

	//if len(sorted.Warn)+len(sorted.Crit) > 0 {
	sendMail(sorted, c.SNSClient, c.SNSTopicArn)
	//}
}

func sendMail(sorted Sorted, snsClient snsiface.SNSAPI, snsTopicArn string) {
	_, e := snsClient.Publish(&sns.PublishInput{
		TopicArn: aws.String(snsTopicArn),
		Message:  aws.String(fmt.Sprintf("%v+", sorted)),
	})

	if e != nil {
		log.WithField("error", e).Fatalf("failed to publish message to topic '%s'", snsTopicArn)
	}
}

func Handler() {
	config.fetchItems()
}

func main() {
	lambda.Start(Handler)
}

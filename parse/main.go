package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
)

type Item struct {
	bibNum  string
	title   string
	details string
	dueDate string
}

type Config struct {
	BaseUrl       string
	Username      string
	Password      string
	Client        HttpClientInterface
	DynamoDBTable string
	DBClient      dynamodbiface.DynamoDBAPI
}

var config = Config{
	BaseUrl:       os.Getenv("BASE_URL"),
	Username:      os.Getenv("USERNAME"),
	Password:      os.Getenv("PASSWORD"),
	DynamoDBTable: os.Getenv("DYNAMODB_TABLE"),
	Client:        &http.Client{},
	DBClient:      dynamodb.New(session.Must(session.NewSession())),
}

type HttpClientInterface interface {
	PostForm(url string, data url.Values) (resp *http.Response, err error)
}

func Handler() {
	dat := config.MustReadWebsiteData()
	items, err := config.ParseItemsFromHtml(dat)
	if err != nil {
		log.WithError(err).Error("Failed to parse html")
	}

	for k, v := range items {
		persist(k, v, config.DBClient)
	}
}

func persist(key string, item Item, dbClient dynamodbiface.DynamoDBAPI) {
	_, err := dbClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(config.DynamoDBTable),
		Item: map[string]*dynamodb.AttributeValue{
			"id":       {S: aws.String(item.bibNum)},
			"title":    {S: aws.String(item.title)},
			"due_date": {S: aws.String(item.dueDate)},
		}})
	if err != nil {
		log.WithField("error", err).Error("Could not store item")
	}
	log.Infof("%s: %v\n", key, item)
}

func (c *Config) ParseItemsFromHtml(htmlData []byte) (map[string]Item, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlData))
	if err != nil {
		log.WithError(err).Error("failed to read html data")
		return nil, err
	}
	items := map[string]Item{}
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		titleElement := s.Find("td.title")
		anchor := titleElement.Find("a.title")
		href, found := anchor.Attr("href")
		if !found {
			return
		}
		parts := strings.Split(href, "=")
		if len(parts) != 2 {
			log.Errorf("Could not split '%s' by '='", href)
			return
		}
		id := parts[1]
		var item Item
		if _, found := items[id]; !found {
			item = Item{
				bibNum:  id,
				title:   strings.Trim(anchor.Text(), " "),
				details: strings.Trim(s.Find("span.item-details").Text(), " "),
			}
		}

		dateElement := s.Find("td.date_due")
		dateSpan := dateElement.Find("span")
		dateString, found := dateSpan.Attr("title")
		if !found {
			log.Warnf("no title tag found in '%s'", dateSpan.Text())
		}
		item.dueDate = dateString
		items[id] = item

	})

	return items, nil
}

func (c *Config) MustReadWebsiteData() []byte {
	resp, err := c.Client.PostForm(
		fmt.Sprintf("%s/%s", c.BaseUrl, "opac-user.pl"),
		url.Values{
			"password": {c.Password},
			"userid":   {c.Username},
		},
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body
}

func main() {
	lambda.Start(Handler)
}

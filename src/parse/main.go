package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
)

type Item struct {
	bibNum  string
	title   string
	dueDate string
}

type Accounts struct {
	Accounts []Account `json:"accounts"`
}

type Account struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	Accounts      []Account
	AccountName   string
	BaseUrl       string
	Username      string
	Password      string
	Client        HttpClientInterface
	DynamoDBTable string
	DBClient      dynamodbiface.DynamoDBAPI
}

var config = Config{
	BaseUrl:       os.Getenv("BASE_URL"),
	DynamoDBTable: os.Getenv("DYNAMODB_TABLE"),
	Client:        &http.Client{},
	DBClient:      dynamodb.New(session.Must(session.NewSession())),
}

type HttpClientInterface interface {
	PostForm(url string, data url.Values) (resp *http.Response, err error)
}

func Handler() {
	accountNames := strings.Split(os.Getenv("ACCOUNT_NAMES"), ",")
	var accounts []*Account

	for _, accountName := range accountNames {
		account, e := fetchAccount(accountName)
		if e != nil {
			log.WithError(e).Fatal("failed to fetch account '%s'", accountName)
		}
		accounts = append(accounts, account)
	}

	for _, a := range accounts {
		dat := config.mustReadWebsiteData(a.Username, a.Password)
		items, err := config.parseItemsFromHtml(dat)
		if err != nil {
			log.WithError(err).Error("Failed to parse html")
		}

		output, err := config.DBClient.Scan(&dynamodb.ScanInput{TableName: aws.String(config.DynamoDBTable)})
		if err != nil {
			log.WithField("error", err).Fatalf("Failed to scan table '%s'", config.DynamoDBTable)
		}

		for _, i := range output.Items {
			id := *i["id"].S

			if _, hasItem := items[id]; !hasItem {
				key := map[string]*dynamodb.AttributeValue{"id": {S: aws.String(id)}}
				_, err := config.DBClient.DeleteItem(&dynamodb.DeleteItemInput{TableName: aws.String(config.DynamoDBTable), Key: key})
				if err != nil {
					log.WithError(err).Warnf("failed to delete item with key '%s'", id)
				}
			}

		}

		for k, v := range items {
			config.persist(k, a.Name, v)
		}
	}
}

func fetchAccount(accountName string) (*Account, error) {
	usernameId := "/libre/prod/accounts/" + accountName + "/username"
	passwordId := "/libre/prod/accounts/" + accountName + "/password"

	svc := secretsmanager.New(session.Must(session.NewSession()))

	username, err := getSecret(usernameId, svc)
	if err != nil {
		log.WithError(err).Errorf("failed to get username with id '%s'", usernameId)
		return nil, err
	}
	password, err := getSecret(passwordId, svc)
	if err != nil {
		log.WithError(err).Errorf("failed to get password with id '%s'", passwordId)
		return nil, err
	}

	return &Account{
		Name:     accountName,
		Username: username,
		Password: password,
	}, nil
}

func getSecret(secretName string, svc *secretsmanager.SecretsManager) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	result, err := svc.GetSecretValue(input)
	if err != nil {
		log.WithError(err).Errorf("failed to get secret value with id '%s'", secretName)
		return "", err
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret '%s' had no secret string", secretName)
	}
	return *result.SecretString, nil
}

func (c *Config) persist(key, accountName string, item Item) {
	_, err := c.DBClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(config.DynamoDBTable),
		Item: map[string]*dynamodb.AttributeValue{
			"id":       {S: aws.String(item.bibNum)},
			"title":    {S: aws.String(item.title)},
			"due_date": {S: aws.String(item.dueDate)},
			"account":  {S: aws.String(accountName)},
		}})
	if err != nil {
		log.WithField("error", err).Error("Could not store item")
	}
	log.Infof("%s: %v\n", key, item)
}

func (c *Config) parseItemsFromHtml(htmlData []byte) (map[string]Item, error) {
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
				bibNum: id,
				title:  strings.Trim(anchor.Text(), " "),
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

func (c *Config) mustReadWebsiteData(username, password string) []byte {
	resp, err := c.Client.PostForm(
		fmt.Sprintf("%s/%s", c.BaseUrl, "opac-user.pl"),
		url.Values{
			"userid":   {username},
			"password": {password},
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

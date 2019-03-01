package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
)

type Item struct {
	bibNum  string
	title   string
	details string
}

type Config struct {
	BaseUrl  string
	Username string
	Password string
	Client   HttpClientInterface
}

var config = Config{
	BaseUrl:  os.Getenv("BASE_URL"),
	Username: os.Getenv("USERNAME"),
	Password: os.Getenv("PASSWORD"),
	Client:   &http.Client{},
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
		persist(k, v)
	}
}

func persist(key string, item Item) {
	log.Warnf("%s: %v\n", key, item)
}

func (c *Config) ParseItemsFromHtml(htmlData []byte) (map[string]Item, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlData))
	if err != nil {
		log.WithError(err).Error("failed to read html data")
		return nil, err
	}
	items := map[string]Item{}
	doc.Find("td.title").Each(func(i int, s *goquery.Selection) {
		anchor := s.Find("a.title")
		href, found := anchor.Attr("href")
		if !found {
			log.WithError(err).Errorf("No href attribute found in '%s'", anchor.Text())
			return
		}
		parts := strings.Split(href, "=")
		if len(parts) != 2 {
			log.Errorf("Could not split '%s' by '='", href)
			return
		}
		id := parts[1]
		if _, found := items[id]; !found {
			items[id] = Item{
				bibNum:  id,
				title:   strings.Trim(anchor.Text(), " "),
				details: strings.Trim(s.Find("span.item-details").Text(), " "),
			}
		}
	})
	return items, nil
}

func mustReadTestData() []byte {
	dat, err := ioutil.ReadFile("./sample.html")
	if err != nil {
		log.Fatal(err)
	}
	return dat
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

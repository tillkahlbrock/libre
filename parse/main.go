package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

func Handler() {
	dat := mustReadWebsiteData()
	items, err := parseItemsFromHtml(dat)
	if err != nil {
		log.WithError(err).Error("Failed to parse html")
	}

	for _, i := range items {
		fmt.Printf("%v\n", i)
	}
}

func parseItemsFromHtml(htmlData []byte) (map[string]Item, error) {
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
			err := fmt.Errorf("some error")
			log.WithError(err).Infof("Found new item with id: '%s'", id)
			items[id] = Item{
				bibNum:  id,
				title:   strings.Trim(anchor.Text(), " "),
				details: strings.Trim(s.Find("span.item-details").Text(), " "),
			}
		}
	})
	return items, nil
}

func mustReadWebsiteData() []byte {
	dat, err := ioutil.ReadFile("./sample.html")
	if err != nil {
		log.Fatal(err)
	}
	return dat
}

func main() {
	lambda.Start(Handler)
}

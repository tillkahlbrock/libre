package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/lambda"
)

type Item struct {
	bibNum  string
	title   string
	details string
}

func Handler() {
	dat, err := ioutil.ReadFile("./sample.html")
	if err != nil {
		log.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(dat))
	if err != nil {
		log.Fatal(err)
	}

	items := map[string]Item{}

	doc.Find("td.title").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		anchor := s.Find("a.title")
		href, found := anchor.Attr("href")
		if !found {
			log.Fatal("No href attribute found")
		}
		parts := strings.Split(href, "=")
		if len(parts) != 2 {
			log.Fatalf("Could not split '%s' by '='", href)
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

	for _, i := range items {
		fmt.Printf("%v\n", i)
	}
}

func main() {
	lambda.Start(Handler)
}

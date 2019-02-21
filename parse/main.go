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

func Handler() {
	dat, err := ioutil.ReadFile("./sample.html")
	if err != nil {
		log.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(dat))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("td.title").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		title := strings.Trim(s.Find("a.title").Text(), " ")
		details := strings.Trim(s.Find("span.item-details").Text(), " ")
		fmt.Printf("Item %d: %s - %s\n", i, title, details)
	})
	fmt.Printf("got scheduled")
}

func main() {
	lambda.Start(Handler)
}

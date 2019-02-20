package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

func Handler() {
	fmt.Printf("got scheduled")
}

func main() {
	lambda.Start(Handler)
}

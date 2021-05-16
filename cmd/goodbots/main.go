package main

import (
	"context"
	"log"
	"os"

	bots "github.com/eywu/goodbots"
)

var (
	concurrency int64
)

func main() {
	concurrency = 10
	//err := bots.ResolveNames(concurrency, context.Background(), os.Stdin, os.Stdout)
	err := bots.GoodBots(concurrency, context.Background(), os.Stdin, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

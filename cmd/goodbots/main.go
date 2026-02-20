package main

import (
	"context"
	"flag"
	"log"
	"os"

	bots "github.com/eywu/goodbots"
)

func main() {
	concurrency := flag.Int64("c", 50, "concurrency limit (number of parallel DNS lookups)")
	mode := flag.String("mode", "goodbots", "mode: \"goodbots\" to verify bot IPs, \"resolve\" to resolve all IPs")
	flag.Parse()

	var err error
	switch *mode {
	case "resolve":
		err = bots.ResolveNames(*concurrency, context.Background(), os.Stdin, os.Stdout)
	default:
		err = bots.GoodBots(*concurrency, context.Background(), os.Stdin, os.Stdout)
	}
	if err != nil {
		log.Fatal(err)
	}
}

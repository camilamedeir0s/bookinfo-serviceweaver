package main

import (
	"context"
	"flag"
	"log"

	"github.com/ServiceWeaver/weaver"
	"github.com/camilamedeir0s/bookinfo-serviceweaver/productpage"
)

//go:generate weaver generate ./...

func main() {
	flag.Parse()
	if err := weaver.Run(context.Background(), productpage.Serve); err != nil {
		log.Fatal(err)
	}
}

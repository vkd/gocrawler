package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/vkd/gocrawler"
)

func main() {
	var url string
	if len(os.Args) > 1 {
		url = os.Args[1]
	} else {
		fmt.Println("URL as first arg is required")
		return
	}

	var c gocrawler.Crawler
	httpClient := &http.Client{
		Timeout: 3 * time.Second,
	}
	clientThrottling := &gocrawler.ClientThrottling{
		Client:         httpClient,
		CallsPerPeriod: 3,
	}
	clientThrottling.StartThrottler()

	c.Client = clientThrottling
	c.OnFindNewLinkFunc = func(link string) {
		fmt.Printf("%v\n", link)
	}
	c.LogErrorFunc = func(err error) {
		fmt.Printf("Error: %v\n", err)
	}
	err := c.Crawl(url)
	if err != nil {
		fmt.Printf("Error on crawl: %v\n", err)
	}
	c.Wait()
}

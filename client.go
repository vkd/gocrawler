package gocrawler

import (
	"net/http"
	"time"
)

// Client - interface of client
type Client interface {
	Get(url string) (*http.Response, error)
}

var _ Client = (*http.Client)(nil)

// ClientFunc - func that implemented Client interface
type ClientFunc func(url string) (*http.Response, error)

// Get - run that func
func (c ClientFunc) Get(url string) (*http.Response, error) {
	return c(url)
}

// ClientThrottling - throttling to count call client per duration
type ClientThrottling struct {
	Client Client

	CallsPerPeriod int
	PeriodDuration time.Duration

	throttling chan struct{}
}

// Get - implement Client interface
func (c *ClientThrottling) Get(url string) (*http.Response, error) {
	<-c.throttling
	return c.Client.Get(url)
}

// StartThrottler - start throttler worker
func (c *ClientThrottling) StartThrottler() {
	if c.CallsPerPeriod <= 0 {
		c.CallsPerPeriod = 1000
	}
	if c.PeriodDuration == 0 {
		c.PeriodDuration = time.Second
	}
	c.throttling = make(chan struct{}, c.CallsPerPeriod)
	go func() {
		for {
			c.initializeThrottler()
			time.Sleep(c.PeriodDuration)
		}
	}()
}

func (c *ClientThrottling) initializeThrottler() {
	for i := 0; i < c.CallsPerPeriod; i++ {
		select {
		case c.throttling <- struct{}{}:
		default:
			return
		}
	}
}

package gocrawler

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Crawler - web crawler
type Crawler struct {
	VisitStorage VisitStorager
	Client       Client

	OnFindNewLinkFunc func(link string)
	LinkFilterFunc    func(u *url.URL) bool

	LogErrorFunc func(err error)

	wg     sync.WaitGroup
	domain string
}

// ErrTimeoutIsExpired - wait more then timeout
var ErrTimeoutIsExpired = errors.New("timeout is expired")

// ErrClientIsNil - client is nil
var ErrClientIsNil = errors.New("client is nil")

// ErrWrongStatus - wrong status from client
// TODO: need to expand with code and body
var ErrWrongStatus = errors.New("wrong status")

// Crawl - crawl url
func (c *Crawler) Crawl(u string) error {
	if c == nil {
		c = &Crawler{}
	}
	if c.VisitStorage == nil {
		c.VisitStorage = NewVisitedStorageMapMutex()
	}
	if c.Client == nil {
		return ErrClientIsNil
	}
	if c.LinkFilterFunc == nil {
		c.LinkFilterFunc = c.isSelfDomain
	}

	// parse domain
	if c.domain == "" {
		var err error
		switch {
		case strings.HasPrefix(u, "http://"):
		case strings.HasPrefix(u, "https://"):
		default:
			u = "http://" + u
		}
		parsedURL, err := url.Parse(u)
		if err != nil {
			return err
		}
		c.domain = parsedURL.Host
	}
	c.startCrawl(u)
	return nil
}

// Wait - wait inlimited
func (c *Crawler) Wait() {
	c.wg.Wait()
}

// WaitTimeout - wait to complete crawled
func (c *Crawler) WaitTimeout(timeout time.Duration) error {
	var stop = make(chan struct{})
	go c.waitChan(stop)

	select {
	case <-stop:
	case <-time.NewTimer(timeout).C:
		return ErrTimeoutIsExpired
	}
	return nil
}

func (c *Crawler) waitChan(stop chan<- struct{}) {
	c.wg.Wait()
	stop <- struct{}{}
}

func (c *Crawler) startCrawl(url string) {
	u, err := c.formatURL(url)
	if err != nil {
		c.logerror(err)
		return
	}

	url = u.String()

	if c.VisitStorage == nil || !c.VisitStorage.IsFisrtVisit(url) {
		return
	}
	if c.LinkFilterFunc != nil && !c.LinkFilterFunc(u) {
		return
	}

	c.wg.Add(1)
	go func() {
		c.logerror(c.crawl(url))
		c.wg.Done()
	}()
}

func (c *Crawler) logerror(err error) {
	if err != nil {
		if c.LogErrorFunc != nil {
			c.LogErrorFunc(err)
		}
	}
}

func (c *Crawler) crawl(url string) error {
	body, err := c.visit(url)
	if err != nil {
		return err
	}
	if body == nil {
		return nil
	}

	if c.OnFindNewLinkFunc != nil {
		c.OnFindNewLinkFunc(url)
	}

	err = forEveryLink(body, c.startCrawl)
	if err != nil {
		return err
	}

	err = body.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *Crawler) visit(url string) (io.ReadCloser, error) {
	if c.Client == nil {
		return nil, ErrClientIsNil
	}
	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		return resp.Body, nil
	}
	if resp.StatusCode == http.StatusMovedPermanently {
		loc, err := resp.Location()
		if err != nil {
			return nil, err
		}
		c.startCrawl(loc.String())
		return nil, nil
	}
	return nil, ErrWrongStatus // TODO: need to expand error with status and body
}

func (c *Crawler) formatURL(s string) (*url.URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		u.Host = c.domain
	}
	return u, nil
}

func (c *Crawler) isSelfDomain(u *url.URL) bool {
	return u.Host == c.domain
}

func forEveryLink(r io.Reader, fn func(string)) error {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}
	// TODO: base [href]
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			for _, a := range n.Attr {
				if a.Key == "href" {
					fn(a.Val)
				}
			}
		}
	})
	return nil
}

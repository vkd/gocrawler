package gocrawler

import (
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func Test_forEveryLink(t *testing.T) {
	testLink := "hello.com/index.html"
	r := strings.NewReader("<a href=\"" + testLink + "\" />")
	forEveryLink(r, func(link string) {
		if link != testLink {
			t.Errorf("Wrong link: %v", link)
		}
	})
}

type testClientFunc func(url string) (*http.Response, error)

func (t testClientFunc) Get(url string) (*http.Response, error) {
	return t(url)
}

func TestWalkAllLinksOnlyOnce(t *testing.T) {
	pageIndex := `
		<a href="http://domain.com/1.html" />
		<a href="http://domain.com/2.html" />
		<a href="http://domain.com/3.html" />
	`
	page1 := `
		<a href="http://domain.com/index.html" />
		<a href="http://domain.com/2.html" />
		<a href="http://domain.com/3.html" />

		<a href="http://domain.com/cat.html" />
	`
	page2 := `
		<a href="http://domain.com/1.html" />
		<a href="http://domain.com/index.html" />
		<a href="http://domain.com/3.html" />

		<a href="http://domain.com/cat.html" />
		<a href="http://domain.com/dog.html" />
	`
	page3 := `
		<a href="http://domain.com/1.html" />
		<a href="http://domain.com/2.html" />
		<a href="http://domain.com/index.html" />

		<a href="http://domain.com/about.html" />
		<a href="http://myspace.com/sponsor.html" />
	`

	site := map[string]string{
		"index.html": pageIndex,
		"1.html":     page1,
		"2.html":     page2,
		"3.html":     page3,
	}

	downloaded := map[string]int{}
	var mxDownloaded sync.Mutex

	client := testClientFunc(func(url string) (*http.Response, error) {
		mxDownloaded.Lock()
		downloaded[url]++
		mxDownloaded.Unlock()

		innerURL := strings.TrimPrefix(url, "http://domain.com/")

		// is redirect
		if innerURL == "" {
			resp := &http.Response{
				StatusCode: 301,
				Header:     http.Header{},
			}
			resp.Header.Set("Location", "http://domain.com/index.html")
			return resp, nil
		}

		respBody := site[innerURL]
		return &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(respBody)),
			StatusCode: 200,
		}, nil
	})

	visited := map[string]int{}
	var mxVisited sync.Mutex

	c := Crawler{
		Client:       client,
		VisitStorage: NewVisitedStorageMapMutex(),
	}
	c.OnFindNewLinkFunc = func(link string) {
		mxVisited.Lock()
		visited[link]++
		mxVisited.Unlock()
	}
	err := c.Crawl("http://domain.com/")
	if err != nil {
		t.Errorf("Error on crawl: %v", err)
	}
	err = c.WaitTimeout(time.Second)
	if err != nil {
		t.Errorf("Error on wait: %v", err)
	}

	// asserts
	if len(downloaded) != 8 {
		t.Errorf("Downloaded: %#v", downloaded)
		t.Errorf("Wrong count of downloaded sites: %v", len(downloaded))
	}
	for l, c := range downloaded {
		if c != 1 {
			t.Errorf("Downloaded: %#v", downloaded)
			t.Errorf("Not once downloaded link: %v (all link must be downloaded only by once)", l)
		}
	}

	if len(visited) != 7 {
		t.Errorf("Wrong count of visited sites: %v", len(visited))
	}
	for l, c := range visited {
		if c != 1 {
			t.Errorf("Not once visited link: %v (all link must be visited only by once)", l)
		}
	}
}

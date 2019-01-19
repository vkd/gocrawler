.PHONY: test build clean docker_build docker run
.DEFAULT_GOAL := docker

test:
	go test ./...

build:
	GOOS=linux GOARCH=amd64 go build -o gocrawler cmd/gocrawler/main.go

clean:
	rm gocrawler

docker_build: build
	docker build -t test-crawler .

docker: docker_build clean

run:
	docker run test-crawler

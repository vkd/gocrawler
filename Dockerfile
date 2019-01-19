FROM alpine
RUN apk update && apk add ca-certificates

ADD gocrawler /
ENTRYPOINT ["/gocrawler"]

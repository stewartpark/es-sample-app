FROM golang:1.15 AS builder
MAINTAINER Stewart Park <hello@stewartjpark.com>

WORKDIR /go/src/app
COPY . .
RUN go build

FROM golang:1.15
COPY --from=builder /go/src/app/es-sample-app /
ENTRYPOINT ["/es-sample-app"]

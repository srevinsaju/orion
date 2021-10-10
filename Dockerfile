# https://chemidy.medium.com/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
FROM golang:alpine AS builder


# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git


WORKDIR $GOPATH/src/github.com/srevinsaju/orion


COPY . .
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /go/bin/orion


FROM alpine:latest
COPY --from=builder /go/bin/orion /go/bin/orion
ENTRYPOINT ["/go/bin/orion", "/etc/orion/config.json"]

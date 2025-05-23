GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

all: lint test build

lint:
	golangci-lint run ./...

test:
	go test -race ./...

test-coverage:
	go test -race -covermode=atomic -coverprofile=coverage.txt ./...

build:
	CGO_ENABLED=0 go build -o ./bin/regauth-$(GOOS)-$(GOARCH) ./cmd/regauth
	CGO_ENABLED=0 go build -o ./bin/regauth-cli-$(GOOS)-$(GOARCH) ./cmd/regauth-cli

docker:
	docker build -t localhost/evanebb/regauth:latest .

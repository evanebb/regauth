GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

build:
	CGO_ENABLED=0 go build -tags viper_bind_struct -o ./bin/regauth-$(GOOS)-$(GOARCH) ./cmd/regauth
	CGO_ENABLED=0 go build -o ./bin/regauth-cli-$(GOOS)-$(GOARCH) ./cmd/regauth-cli

test:
	go test -race ./...

test-coverage:
	go test -race -covermode=atomic -coverprofile=coverage.txt ./...

docker:
	docker build -t localhost/evanebb/regauth:latest .

build:
	CGO_ENABLED=0 go build -tags viper_bind_struct -o ./bin/regauth ./cmd/regauth
	CGO_ENABLED=0 go build -o ./bin/regauth-cli ./cmd/regauth-cli

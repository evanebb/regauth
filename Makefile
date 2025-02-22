build:
	CGO_ENABLED=0 go build -tags viper_bind_struct -o ./bin/regauth ./cmd/regauth

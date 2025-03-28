# Contributing
For development purposes, this repository contains a [docker-compose.yml](docker-compose.yml) that can be used for testing locally.

## Generate signing key and certificate
```shell
openssl genrsa -out key.pem 4096
openssl req -new -key key.pem -x509 -nodes -days 365 -out cert.pem
```

## Build
To build the binaries:
```shell
make build
```

This will result in two binaries:
- `bin/regauth-$OS-$ARCH`, the server
- `bin/regauth-cli-$OS-$ARCH`, the CLI

The names of both binaries depend on the OS/architecture of your machine, so for x86 Linux it would be `regauth-linux-amd64`, and for an ARM Mac it would be `regauth-darwin-arm64`.

To build the Docker image:
```shell
make docker
```

## Run
To run a setup using the locally-built binary:
```shell
docker compose up -d
./bin/regauth-$OS-$ARCH serve configuration/config-dev.yml
```

Or to use the locally-built Docker image:
```shell
docker compose --profile regauth up -d
```

You can then log into the instance and generate a new personal access token with:
```shell
./bin/regauth-cli-$OS-$ARCH login https://localhost:8000 --username admin --password admin
regauth-cli-$OS-$ARCH token create --description test-token --expirationDate 2030-01-01T00:00:00Z --permission readWriteDelete --login
```

The API reference can now also be found at [http://localhost:8000](http://localhost:8000).

## Test
To run the unit tests:
```shell
make test
```

## Mock data
This repository also contains mock data/seeds that can be used, although they are mostly meant for the automated tests.
To use them:
```shell
goose -dir ./resources/database/seeds postgres "postgres://regauth:Welkom01@127.0.0.1:5432/regauth" --no-versioning up
```

[Goose](https://github.com/pressly/goose) is used for the database migrations/seeds, so you need to make sure that it is installed first.

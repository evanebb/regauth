FROM golang:1.23-alpine AS build

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -tags viper_bind_struct -o ./bin/regauth ./cmd/regauth

FROM scratch

WORKDIR /etc/regauth
COPY configuration/config-docker.yml ./config.yml
COPY --from=build /app/bin/regauth /regauth

CMD ["/regauth", "serve", "/etc/regauth/config.yml"]

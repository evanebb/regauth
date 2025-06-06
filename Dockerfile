FROM golang:1.24.0-alpine AS build

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o ./bin/regauth ./cmd/regauth

FROM scratch

WORKDIR /etc/regauth
COPY configuration/config-docker.yml ./config.yml
COPY --from=build /app/bin/regauth /regauth

USER 65532:65532

CMD ["/regauth", "serve", "/etc/regauth/config.yml"]

# Docker Compose instructions

## Generate signing key and certificate

```shell
openssl genrsa -out key.pem 4096
openssl req -new -key key.pem -x509 -nodes -days 365 -out cert.pem
```

The above commands will generate an RSA key/certificate, to be used with the RS256 JWT signing algorithm.
This application supports all asymmetric JWT signing algorithms as of the time of writing.
However, if you want to use a different algorithm, you must change the `token.alg` option in
the [regauth-config.yml](regauth-config.yml) file accordingly.

## Change the defaults

You should change the default admin user and database passwords inside the [docker-compose.yml](docker-compose.yml) file
to something more secure.

For the `regauth` container, change the following environment variable values:

- `REGAUTH_INITIALADMIN_USERNAME`
- `REGAUTH_INITIALADMIN_PASSWORD`
- `REGAUTH_DATABASE_PASSWORD`

For the `postgres` container:

- `POSTGRES_PASSWORD` (same value as the `REGAUTH_DATABASE_PASSWORD` variable!)

You should also change the `auth.token.realm` option in [registry-config.yml](registry-config.yml) to the externally reachable URL of your regauth instance.

## Run the application

```shell
docker compose up -d
```

The application can now be reached at `http://localhost:8000`.

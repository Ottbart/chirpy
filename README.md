# Chirpy

Chirpy is a Go HTTP API for creating and reading short posts called chirps. It stores users, chirps, refresh tokens, and Chirpy Red membership in PostgreSQL. The API supports password-based login, JWT authentication, and refresh-token revocation.

## Requirements

- Go 1.26.5 or later
- PostgreSQL
- [Goose](https://github.com/pressly/goose) CLI for applying database migrations

## Configuration

The server loads a `.env` file from the project directory and requires these values:

| Variable | Purpose |
| --- | --- |
| `DB_URL` | PostgreSQL connection URL |
| `PLATFORM` | Runtime platform; set to `dev` to enable the database reset endpoint |
| `SECRET` | Secret used to sign access JWTs |
| `POLKA_KEY` | API key used to authenticate Polka webhooks |

For example, create a local `.env` file with values appropriate for your PostgreSQL setup:

```dotenv
DB_URL=postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
SECRET=replace-with-a-long-random-secret
POLKA_KEY=replace-with-your-polka-api-key
```

Keep real secrets out of source control.

## Database Setup

Start PostgreSQL and create the database referenced by `DB_URL` (the example URL uses a database named `chirpy`). A connection URL is a value for `DB_URL`, not a command to run. Load the `.env` variables into your shell, then apply the schema migrations from `sql/schema`:

```sh
set -a
. ./.env
set +a
goose -dir sql/schema postgres "$DB_URL" up
```

The SQL queries are in `sql/queries`; generated database access code is in `internal/database` and is configured by `sqlc.yaml`.

## Start Everything

Before the first run, configure `.env`, create the `chirpy` database, and apply the migrations using the steps in [Database Setup](#database-setup). Then, from the project directory:

1. Start the PostgreSQL service:

  ```sh
  sudo systemctl start postgresql
  ```

2. Connect to the database to verify access. Replace `username` and `password` with your PostgreSQL credentials. At the `psql` prompt, enter `\q` to exit before continuing:

  ```sh
  psql "postgres://username:password@localhost:5432/chirpy"
  ```

3. Start the Chirpy API:

  ```sh
  go run .
  ```

The API listens on port `8080`. Leave it running while you use the API or run the Boot.dev lesson in another terminal.

## Run

From the project directory, start the API with:

```sh
go run .
```

The HTTP server listens on port `8080`. Static project files are served under `/app/`.

Run the test suite with:

```sh
go test ./...
```

## API

JSON request bodies should use `Content-Type: application/json`. Authenticated routes expect an `Authorization: Bearer <token>` header.

| Method and path | Description | Authentication |
| --- | --- | --- |
| `GET /api/healthz` | Readiness check | None |
| `POST /api/users` | Create an account with `email` and `password` | None |
| `PUT /api/users` | Update the current user's email and password | Access JWT |
| `POST /api/login` | Log in with `email` and `password`; returns an access JWT and refresh token | None |
| `POST /api/refresh` | Exchange a valid refresh token for a new access JWT | Refresh token |
| `POST /api/revoke` | Revoke a refresh token | Refresh token |
| `POST /api/chirps/` | Create a chirp with a `body` | Access JWT |
| `GET /api/chirps/` | List chirps; supports optional `author_id` and `sort=asc` or `sort=desc` query parameters | None |
| `GET /api/chirps/{chirpID}` | Get one chirp | None |
| `DELETE /api/chirps/{chirpID}` | Delete a chirp owned by the authenticated user | Access JWT |
| `POST /api/polka/webhooks` | Process a Polka user-upgrade event | Polka API key |
| `GET /admin/metrics` | Show the static file request count | None |
| `POST /admin/reset` | Reset the request count and delete users | `PLATFORM=dev` only |

Chirps are limited to 140 characters. The words `kerfuffle`, `sharbert`, and `fornax` are replaced with asterisks when a chirp is created.

### Example: Create an account and log in

```sh
curl -X POST http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"reader@example.com","password":"change-me"}'

curl -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"reader@example.com","password":"change-me"}'
```

The login response includes `token` and `refresh_token`. Use the access token for protected user and chirp routes. Use the refresh token as a Bearer token with `/api/refresh` or `/api/revoke`.

### Example: Create a chirp

```sh
curl -X POST http://localhost:8080/api/chirps/ \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <access-token>' \
  -d '{"body":"Hello from Chirpy!"}'
```
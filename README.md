# PostgreSQL MCP

An MCP server for PostgreSQL.

## Run

All PostgreSQL connection flags are required:

```sh
go run . \
  --host localhost \
  --port 5432 \
  --user postgres \
  --password secret \
  --database app
```

The default transport is stdio. To run the streamable HTTP transport:

```sh
go run . \
  --transport http \
  --mcpaddr :8080 \
  --host localhost \
  --port 5432 \
  --user postgres \
  --password secret \
  --database app
```

Use `--sslmode` to change the PostgreSQL SSL mode. It defaults to `disable` for
simple local PostgreSQL setups. The server pings PostgreSQL before starting and
the MCP `ping` tool checks the connection again when called. The MCP `query`
tool accepts a SQL string through its `sql` argument and returns the result
columns and rows.

## Integration Tests

The `ping` and `query` integration tests connect directly to PostgreSQL. They
are opt-in and use the `docker-compose.yml` connection values by default:

```sh
docker compose up -d
TEST_POSTGRES_INTEGRATION=1 go test ./...
```

Adjust the test connection with `TEST_POSTGRES_DSN`, or with these individual
variables: `TEST_POSTGRES_HOST`, `TEST_POSTGRES_PORT`, `TEST_POSTGRES_USER`,
`TEST_POSTGRES_PASSWORD`, `TEST_POSTGRES_DATABASE`, and `TEST_POSTGRES_SSLMODE`.
For example:

```sh
TEST_POSTGRES_INTEGRATION=1 \
TEST_POSTGRES_HOST=127.0.0.1 \
TEST_POSTGRES_PORT=5432 \
TEST_POSTGRES_USER=postgres \
TEST_POSTGRES_PASSWORD=secret \
TEST_POSTGRES_DATABASE=app \
go test ./...
```

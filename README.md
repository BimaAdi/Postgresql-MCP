# PostgreSQL MCP

An MCP server for PostgreSQL.

## How to use
### Run as executable (Recomended)
1. Get the executable
there are 2 options:
- download the latest release see [releases](https://github.com/BimaAdi/Postgresql-MCP/releases)
- build from source see [build from source](./docs/build_from_source.md)

2. Integrate it with your code agent see [code agent integration](./docs/code_agent_integration.md)

### Run Manually

Provide the PostgreSQL connection flags on the command line:

```sh
go run ./cmd \
  --host localhost \
  --port 5432 \
  --user postgres \
  --password secret \
  --database app
```

Each flag can also be provided through an environment variable when the flag is
not set. Command-line flags take precedence:

| Flag | Environment variable |
| --- | --- |
| `--host` | `POSTGRES_HOST` |
| `--port` | `POSTGRES_PORT` |
| `--user` | `POSTGRES_USER` |
| `--password` | `POSTGRES_PASSWORD` |
| `--database` | `POSTGRES_DATABASE` |
| `--sslmode` | `POSTGRES_SSLMODE` |
| `--transport` | `MCP_TRANSPORT` |
| `--mcpaddr` | `MCP_ADDR` |

For example:

```sh
POSTGRES_HOST=localhost \
POSTGRES_PORT=5432 \
POSTGRES_USER=postgres \
POSTGRES_PASSWORD=secret \
POSTGRES_DATABASE=app \
go run ./cmd
```

The default transport is stdio. To run the streamable HTTP transport:

```sh
go run ./cmd \
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

Next integrate it with your code agent see [code agent integration](./docs/code_agent_integration.md)
but change it to run go instead of executable

### As a Library
The root package is also importable by Go applications. It accepts an existing
`*sql.DB`, so your application is responsible for opening, configuring, and
closing the database connection. The MCP SDK is responsible for running the
server over a transport.

Create a new Go application and add the required dependencies:

```sh
mkdir postgres-mcp-app
cd postgres-mcp-app
go mod init example.com/postgres-mcp-app
go get github.com/BimaAdi/postgresql-mcp github.com/lib/pq github.com/modelcontextprotocol/go-sdk
```

The following complete program reads a PostgreSQL URL from `DATABASE_URL`,
checks the connection, creates the MCP server, and runs it over stdio:

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	postgresqlmcp "github.com/BimaAdi/postgresql-mcp"
	_ "github.com/lib/pq"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open PostgreSQL connection: %w", err)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// sql.Open does not establish a connection until the database is used.
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	server := postgresqlmcp.NewServer(db)
	return server.Run(ctx, &mcp.StdioTransport{})
}

```

Run it with a PostgreSQL connection URL:

```sh
DATABASE_URL='postgres://postgres:secret@localhost:5432/app?sslmode=disable' go run .
```

For a URL containing special characters in the username or password, percent-
encode those values, or construct the URL with a PostgreSQL URL helper before
calling `sql.Open`. Do not commit credentials to source control.

To expose the same server over streamable HTTP instead of stdio, replace the
last two lines of `run` with:

```go
handler := mcp.NewStreamableHTTPHandler(
	func(*http.Request) *mcp.Server { return server },
	&mcp.StreamableHTTPOptions{},
)
httpServer := &http.Server{Addr: ":8080", Handler: handler}
go func() {
	<-ctx.Done()
	_ = httpServer.Shutdown(context.Background())
}()
if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	return fmt.Errorf("serve HTTP: %w", err)
}
return nil
```

The HTTP variant also needs these imports:

```go
"errors"
"net/http"
```

Keep the `*sql.DB` open for the lifetime of the MCP server and close it when
the server exits. `database/sql` manages a connection pool; configure pool
limits such as `db.SetMaxOpenConns` in your application when the defaults are
not suitable. The server provides the `ping` and `query` tools, and `query`
executes the SQL supplied by the MCP client, so only expose it to trusted
clients.

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

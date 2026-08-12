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

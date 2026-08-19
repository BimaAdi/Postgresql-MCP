package postgresqlmcp

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type pingInput struct{}

// PingOutput is the structured output returned by the ping tool.
type PingOutput struct {
	Connected bool   `json:"connected"`
	Message   string `json:"message"`
}

type queryInput struct {
	SQL string `json:"sql" jsonschema:"SQL query to execute"`
}

// QueryOutput is the structured output returned by the query tool.
type QueryOutput struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

// NewServer creates an MCP server that uses db for PostgreSQL operations.
func NewServer(db *sql.DB) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "postgresql-mcp", Version: "0.1.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ping",
		Description: "Check whether the PostgreSQL database connection is working",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ pingInput) (*mcp.CallToolResult, PingOutput, error) {
		if err := db.PingContext(ctx); err != nil {
			return nil, PingOutput{Connected: false, Message: "PostgreSQL connection failed"}, fmt.Errorf("PostgreSQL connection failed: %w", err)
		}
		return nil, PingOutput{Connected: true, Message: "PostgreSQL connection is healthy"}, nil
	})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "query",
		Description: "Execute a SQL query against PostgreSQL and return its columns and rows",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input queryInput) (*mcp.CallToolResult, QueryOutput, error) {
		rows, err := db.QueryContext(ctx, input.SQL)
		if err != nil {
			return nil, QueryOutput{}, fmt.Errorf("execute SQL query: %w", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, QueryOutput{}, fmt.Errorf("read query columns: %w", err)
		}
		output := QueryOutput{Columns: columns, Rows: make([][]any, 0)}
		for rows.Next() {
			values := make([]any, len(columns))
			scanTargets := make([]any, len(columns))
			for i := range values {
				scanTargets[i] = &values[i]
			}
			if err := rows.Scan(scanTargets...); err != nil {
				return nil, QueryOutput{}, fmt.Errorf("read query row: %w", err)
			}
			output.Rows = append(output.Rows, values)
		}
		if err := rows.Err(); err != nil {
			return nil, QueryOutput{}, fmt.Errorf("iterate query rows: %w", err)
		}
		return nil, output, nil
	})
	return server
}

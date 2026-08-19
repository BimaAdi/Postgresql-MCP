package postgresqlmcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPostgreSQLTools(t *testing.T) {
	if os.Getenv("TEST_POSTGRES_INTEGRATION") != "1" {
		t.Skip("set TEST_POSTGRES_INTEGRATION=1 to run PostgreSQL integration tests")
	}

	cfg := testPostgreSQLConfig(t)
	dsn := cfg
	if value := os.Getenv("TEST_POSTGRES_DSN"); value != "" {
		dsn = value
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open PostgreSQL connection: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	server := NewServer(db)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("connect MCP server: %v", err)
	}
	t.Cleanup(func() { serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "postgresql-mcp-test", Version: "test"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect MCP client: %v", err)
	}
	t.Cleanup(func() { clientSession.Close() })

	t.Run("ping", func(t *testing.T) {
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "ping"})
		if err != nil {
			t.Fatalf("call ping: %v", err)
		}

		var output PingOutput
		decodeStructuredOutput(t, result, &output)
		if !output.Connected {
			t.Fatalf("ping connected = false, message = %q", output.Message)
		}
		if output.Message != "PostgreSQL connection is healthy" {
			t.Errorf("ping message = %q", output.Message)
		}
	})

	t.Run("query", func(t *testing.T) {
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: "query", Arguments: map[string]any{"sql": "SELECT 7 AS id, 'postgresql-mcp' AS name, true AS enabled"},
		})
		if err != nil {
			t.Fatalf("call query: %v", err)
		}

		var output QueryOutput
		decodeStructuredOutput(t, result, &output)
		wantColumns := []string{"id", "name", "enabled"}
		if len(output.Columns) != len(wantColumns) {
			t.Fatalf("query columns = %v, want %v", output.Columns, wantColumns)
		}
		for i, want := range wantColumns {
			if output.Columns[i] != want {
				t.Errorf("query column %d = %q, want %q", i, output.Columns[i], want)
			}
		}
		if len(output.Rows) != 1 {
			t.Fatalf("query rows = %v, want one row", output.Rows)
		}
		if got, want := output.Rows[0], []any{int64(7), "postgresql-mcp", true}; !equalJSONValues(got, want) {
			t.Errorf("query row = %v, want %v", got, want)
		}
	})
}

func testPostgreSQLConfig(t *testing.T) string {
	t.Helper()
	port := 5434
	if value := os.Getenv("TEST_POSTGRES_PORT"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 65535 {
			t.Fatalf("invalid TEST_POSTGRES_PORT %q", value)
		}
		port = parsed
	}
	return "postgres://" + envOrDefault("TEST_POSTGRES_USER", "user") + ":" + envOrDefault("TEST_POSTGRES_PASSWORD", "password") + "@" + envOrDefault("TEST_POSTGRES_HOST", "localhost") + ":" + strconv.Itoa(port) + "/" + envOrDefault("TEST_POSTGRES_DATABASE", "mcp") + "?sslmode=" + envOrDefault("TEST_POSTGRES_SSLMODE", "disable")
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func decodeStructuredOutput(t *testing.T, result *mcp.CallToolResult, output any) {
	t.Helper()
	if result.IsError {
		t.Fatalf("tool returned an error: %v", result.Content)
	}
	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured output: %v", err)
	}
	if err := json.Unmarshal(data, output); err != nil {
		t.Fatalf("decode structured output: %v", err)
	}
}

func equalJSONValues(got, want any) bool {
	gotJSON, err := json.Marshal(got)
	if err != nil {
		return false
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		return false
	}
	return string(gotJSON) == string(wantJSON)
}

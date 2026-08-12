package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type config struct {
	host      string
	port      int
	user      string
	password  string
	database  string
	transport string
	address   string
	sslmode   string
}

type pingInput struct{}

type pingOutput struct {
	Connected bool   `json:"connected"`
	Message   string `json:"message"`
}

type queryInput struct {
	SQL string `json:"sql" jsonschema:"SQL query to execute"`
}

type queryOutput struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", cfg.dsn())
	if err != nil {
		return fmt.Errorf("open PostgreSQL connection: %w", err)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	server := newServer(db)
	switch cfg.transport {
	case "stdio":
		return server.Run(ctx, &mcp.StdioTransport{})
	case "http":
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, &mcp.StreamableHTTPOptions{})
		httpServer := &http.Server{Addr: cfg.address, Handler: handler}
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = httpServer.Shutdown(shutdownCtx)
		}()

		slog.Info("MCP server listening", "address", cfg.address, "transport", "http")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported transport %q", cfg.transport)
	}
}

func parseConfig(args []string) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("postgresql-mcp", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&cfg.host, "host", "", "PostgreSQL host (required)")
	flags.IntVar(&cfg.port, "port", 0, "PostgreSQL port (required)")
	flags.StringVar(&cfg.user, "user", "", "PostgreSQL user (required)")
	flags.StringVar(&cfg.password, "password", "", "PostgreSQL password (required)")
	flags.StringVar(&cfg.database, "database", "", "PostgreSQL database (required)")
	flags.StringVar(&cfg.transport, "transport", "stdio", "MCP transport: stdio or http")
	flags.StringVar(&cfg.address, "mcpaddr", ":8080", "MCP HTTP listen address")
	flags.StringVar(&cfg.sslmode, "sslmode", "disable", "PostgreSQL SSL mode")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	if flags.NArg() != 0 {
		return config{}, fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}

	for name, value := range map[string]string{
		"--host": cfg.host, "--user": cfg.user, "--password": cfg.password, "--database": cfg.database,
	} {
		if value == "" {
			return config{}, fmt.Errorf("%s is required", name)
		}
	}
	if cfg.port < 1 || cfg.port > 65535 {
		return config{}, fmt.Errorf("--port must be between 1 and 65535")
	}
	if cfg.transport != "stdio" && cfg.transport != "http" {
		return config{}, fmt.Errorf("--transport must be stdio or http")
	}
	if cfg.transport == "http" && cfg.address == "" {
		return config{}, fmt.Errorf("--mcpaddr must not be empty for HTTP transport")
	}
	return cfg, nil
}

func (c config) dsn() string {
	u := &url.URL{
		Scheme: "postgres",
		Host:   c.host + ":" + strconv.Itoa(c.port),
		Path:   "/" + c.database,
		User:   url.UserPassword(c.user, c.password),
	}
	query := u.Query()
	query.Set("connect_timeout", "5")
	query.Set("sslmode", c.sslmode)
	u.RawQuery = query.Encode()
	return u.String()
}

func newServer(db *sql.DB) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "postgresql-mcp", Version: "0.1.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ping",
		Description: "Check whether the PostgreSQL database connection is working",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ pingInput) (*mcp.CallToolResult, pingOutput, error) {
		if err := db.PingContext(ctx); err != nil {
			return nil, pingOutput{Connected: false, Message: "PostgreSQL connection failed"}, fmt.Errorf("PostgreSQL connection failed: %w", err)
		}
		return nil, pingOutput{Connected: true, Message: "PostgreSQL connection is healthy"}, nil
	})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "query",
		Description: "Execute a SQL query against PostgreSQL and return its columns and rows",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input queryInput) (*mcp.CallToolResult, queryOutput, error) {
		rows, err := db.QueryContext(ctx, input.SQL)
		if err != nil {
			return nil, queryOutput{}, fmt.Errorf("execute SQL query: %w", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, queryOutput{}, fmt.Errorf("read query columns: %w", err)
		}
		output := queryOutput{Columns: columns, Rows: make([][]any, 0)}
		for rows.Next() {
			values := make([]any, len(columns))
			scanTargets := make([]any, len(columns))
			for i := range values {
				scanTargets[i] = &values[i]
			}
			if err := rows.Scan(scanTargets...); err != nil {
				return nil, queryOutput{}, fmt.Errorf("read query row: %w", err)
			}
			output.Rows = append(output.Rows, values)
		}
		if err := rows.Err(); err != nil {
			return nil, queryOutput{}, fmt.Errorf("iterate query rows: %w", err)
		}
		return nil, output, nil
	})
	return server
}

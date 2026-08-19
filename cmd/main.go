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

	postgresqlmcp "github.com/BimaAdi/postgresql-mcp"
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

	server := postgresqlmcp.NewServer(db)
	switch cfg.transport {
	case "stdio":
		return server.Run(ctx, &mcp.StdioTransport{})
	case "http":
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{})
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

	envValues := map[string]*string{"host": &cfg.host, "user": &cfg.user, "password": &cfg.password, "database": &cfg.database, "transport": &cfg.transport, "mcpaddr": &cfg.address, "sslmode": &cfg.sslmode}
	envNames := map[string]string{"host": "POSTGRES_HOST", "user": "POSTGRES_USER", "password": "POSTGRES_PASSWORD", "database": "POSTGRES_DATABASE", "transport": "MCP_TRANSPORT", "mcpaddr": "MCP_ADDR", "sslmode": "POSTGRES_SSLMODE"}
	for name, value := range envValues {
		if flagWasSet(flags, name) {
			continue
		}
		if envValue := os.Getenv(envNames[name]); envValue != "" {
			*value = envValue
		}
	}
	if !flagWasSet(flags, "port") {
		if envValue := os.Getenv("POSTGRES_PORT"); envValue != "" {
			port, err := strconv.Atoi(envValue)
			if err != nil {
				return config{}, fmt.Errorf("invalid POSTGRES_PORT %q: %w", envValue, err)
			}
			cfg.port = port
		}
	}

	for name, value := range map[string]string{"--host": cfg.host, "--user": cfg.user, "--password": cfg.password, "--database": cfg.database} {
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

func flagWasSet(flags *flag.FlagSet, name string) bool {
	set := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})
	return set
}

func (c config) dsn() string {
	u := &url.URL{Scheme: "postgres", Host: c.host + ":" + strconv.Itoa(c.port), Path: "/" + c.database, User: url.UserPassword(c.user, c.password)}
	query := u.Query()
	query.Set("connect_timeout", "5")
	query.Set("sslmode", c.sslmode)
	u.RawQuery = query.Encode()
	return u.String()
}

package main

import "testing"

func TestParseConfigEnvironment(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "env-host")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_USER", "env-user")
	t.Setenv("POSTGRES_PASSWORD", "env-password")
	t.Setenv("POSTGRES_DATABASE", "env-database")
	t.Setenv("POSTGRES_SSLMODE", "require")
	t.Setenv("MCP_TRANSPORT", "http")
	t.Setenv("MCP_ADDR", ":9090")

	cfg, err := parseConfig(nil)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	want := config{host: "env-host", port: 5433, user: "env-user", password: "env-password", database: "env-database", sslmode: "require", transport: "http", address: ":9090"}
	if cfg != want {
		t.Fatalf("parseConfig() = %+v, want %+v", cfg, want)
	}
}

func TestParseConfigFlagsOverrideEnvironment(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "env-host")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_USER", "env-user")
	t.Setenv("POSTGRES_PASSWORD", "env-password")
	t.Setenv("POSTGRES_DATABASE", "env-database")

	cfg, err := parseConfig([]string{"--host", "flag-host", "--port", "5434", "--user", "flag-user", "--password", "flag-password", "--database", "flag-database"})
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	if cfg.host != "flag-host" || cfg.port != 5434 || cfg.user != "flag-user" || cfg.password != "flag-password" || cfg.database != "flag-database" {
		t.Fatalf("flag values did not override environment: %+v", cfg)
	}
}

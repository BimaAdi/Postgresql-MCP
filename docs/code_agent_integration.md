# Code Agent Integration
## Claude Code

Add a project-scoped server from the project directory. Replace the executable
path and database values as needed:

Command-line arguments:

```sh
claude mcp add --scope project mcp-postgres \
  -- /path/to/your/postgresql-mcp \
  --host localhost \
  --port 5434 \
  --user user \
  --password password \
  --database mcp
```

Environment variables:

```sh
claude mcp add --scope project mcp-postgres \
  --env POSTGRES_HOST=localhost \
  --env POSTGRES_PORT=5434 \
  --env POSTGRES_USER=user \
  --env POSTGRES_PASSWORD=password \
  --env POSTGRES_DATABASE=mcp \
  -- /path/to/your/postgresql-mcp
```

Check that Claude Code can see it:

```sh
claude mcp list
```

The project scope stores the server in the project's Claude configuration. Use
`--scope user` instead if it should be available in every project.

## Cursor

Create `.cursor/mcp.json` in the project, or edit Cursor's global MCP
configuration:

Command-line arguments:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "args": ["--host", "localhost", "--port", "5434", "--user", "user", "--password", "password", "--database", "mcp"]
    }
  }
}
```

Environment variables:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "env": {
        "POSTGRES_HOST": "localhost",
        "POSTGRES_PORT": "5434",
        "POSTGRES_USER": "user",
        "POSTGRES_PASSWORD": "password",
        "POSTGRES_DATABASE": "mcp"
      }
    }
  }
}
```

Restart or reload Cursor, then open the MCP panel and enable `mcp-postgres`.

## Windsurf

Open Windsurf's MCP settings, or edit
`~/.codeium/windsurf/mcp_config.json`, and add this entry:

Command-line arguments:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "args": ["--host", "localhost", "--port", "5434", "--user", "user", "--password", "password", "--database", "mcp"]
    }
  }
}
```

Environment variables:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "env": {
        "POSTGRES_HOST": "localhost",
        "POSTGRES_PORT": "5434",
        "POSTGRES_USER": "user",
        "POSTGRES_PASSWORD": "password",
        "POSTGRES_DATABASE": "mcp"
      }
    }
  }
}
```

If the file already contains other servers, add `mcp-postgres` inside its
existing `mcpServers` object instead of replacing the file. Restart Windsurf
and check the MCP server status.

## VS Code and GitHub Copilot

Create or edit `.vscode/mcp.json` in the workspace:

Command-line arguments:

```json
{
  "servers": {
    "mcp-postgres": {
      "type": "stdio",
      "command": "/path/to/your/postgresql-mcp",
      "args": ["--host", "localhost", "--port", "5434", "--user", "user", "--password", "password", "--database", "mcp"]
    }
  }
}
```

Environment variables:

```json
{
  "servers": {
    "mcp-postgres": {
      "type": "stdio",
      "command": "/path/to/your/postgresql-mcp",
      "env": {
        "POSTGRES_HOST": "localhost",
        "POSTGRES_PORT": "5434",
        "POSTGRES_USER": "user",
        "POSTGRES_PASSWORD": "password",
        "POSTGRES_DATABASE": "mcp"
      }
    }
  }
}
```

Use the **MCP: List Servers** command in the Command Palette to start the
server. GitHub Copilot Chat uses the same VS Code MCP configuration when MCP
support is enabled. This repository includes an equivalent example at
`.vscode/mcp.json`.

## Cline

In Cline, open **MCP Servers**, choose **Configure**, and add the following
server to the `mcpServers` object in the generated configuration:

Command-line arguments:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "args": ["--host", "localhost", "--port", "5434", "--user", "user", "--password", "password", "--database", "mcp"],
      "disabled": false,
      "alwaysAllow": []
    }
  }
}
```

Environment variables:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "args": [],
      "env": {
        "POSTGRES_HOST": "localhost",
        "POSTGRES_PORT": "5434",
        "POSTGRES_USER": "user",
        "POSTGRES_PASSWORD": "password",
        "POSTGRES_DATABASE": "mcp"
      },
      "disabled": false,
      "alwaysAllow": []
    }
  }
}
```

Save the configuration and enable the server in Cline's MCP panel.

## Roo Code

Open Roo Code's MCP settings and add this server to the `mcpServers` object in
the project or global MCP configuration:

Command-line arguments:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "args": ["--host", "localhost", "--port", "5434", "--user", "user", "--password", "password", "--database", "mcp"],
      "disabled": false,
      "alwaysAllow": []
    }
  }
}
```

Environment variables:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "args": [],
      "env": {
        "POSTGRES_HOST": "localhost",
        "POSTGRES_PORT": "5434",
        "POSTGRES_USER": "user",
        "POSTGRES_PASSWORD": "password",
        "POSTGRES_DATABASE": "mcp"
      },
      "disabled": false,
      "alwaysAllow": []
    }
  }
}
```

Restart the MCP connection from Roo Code and confirm that `ping` and `query`
appear in the available tools.

## Gemini CLI

Create or edit `.gemini/settings.json` in the project, or the user-level
Gemini CLI settings file:

Command-line arguments:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "args": ["--host", "localhost", "--port", "5434", "--user", "user", "--password", "password", "--database", "mcp"]
    }
  }
}
```

Environment variables:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "args": [],
      "env": {
        "POSTGRES_HOST": "localhost",
        "POSTGRES_PORT": "5434",
        "POSTGRES_USER": "user",
        "POSTGRES_PASSWORD": "password",
        "POSTGRES_DATABASE": "mcp"
      }
    }
  }
}
```

Start Gemini CLI from the project and inspect the MCP server status or ask it
to call the `ping` tool.

## Claude Desktop

Edit Claude Desktop's `claude_desktop_config.json` and add the server under
`mcpServers`:

Command-line arguments:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "args": ["--host", "localhost", "--port", "5434", "--user", "user", "--password", "password", "--database", "mcp"]
    }
  }
}
```

Environment variables:

```json
{
  "mcpServers": {
    "mcp-postgres": {
      "command": "/path/to/your/postgresql-mcp",
      "args": [],
      "env": {
        "POSTGRES_HOST": "localhost",
        "POSTGRES_PORT": "5434",
        "POSTGRES_USER": "user",
        "POSTGRES_PASSWORD": "password",
        "POSTGRES_DATABASE": "mcp"
      }
    }
  }
}
```

Typical configuration locations are:

- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\\Claude\\claude_desktop_config.json`

Fully quit and reopen Claude Desktop after saving the file.

## OpenCode

OpenCode uses `mcp` rather than `mcpServers`, and its local command is an
array. Add this entry to `opencode.json`:

Command-line arguments:

```json
{
  "mcp": {
    "mcp-postgres": {
      "type": "local",
      "command": [
        "/path/to/your/postgresql-mcp",
        "--host", "localhost",
        "--port", "5434",
        "--user", "user",
        "--password", "password",
        "--database", "mcp"
      ],
      "enabled": true
    }
  }
}
```

Environment variables:

```json
{
  "mcp": {
    "mcp-postgres": {
      "type": "local",
      "command": [
        "/path/to/your/postgresql-mcp"
      ],
      "environment": {
        "POSTGRES_HOST": "localhost",
        "POSTGRES_PORT": "5434",
        "POSTGRES_USER": "user",
        "POSTGRES_PASSWORD": "password",
        "POSTGRES_DATABASE": "mcp"
      },
      "enabled": true
    }
  }
}
```

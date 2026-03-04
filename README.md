# TABULARIS-REDIS-PLUGIN

A Redis driver plugin for [Tabularis](https://github.com/debba/tabularis), the lightweight database management tool.

This plugin enables Tabularis to connect to Redis databases and explore its data in a tabular format, providing key scanning and virtual table representations for complex data types through a JSON-RPC 2.0 over stdio interface.

## TABLE OF CONTENTS

* Features
* Supported Redis Data Types
* Installation
    * Automatic (via Tabularis)
    * Manual Installation
* How It Works
* Virtual Tables
* Supported Operations
* Building from Source
* Development
* License

## FEATURES

* **Key Scanning** — View all keys with their types and Time-To-Live (TTL).
* **Virtual Tables** — Explore Hashes, Lists, Sets, and Sorted Sets as virtual relational tables.
* **JSON-RPC 2.0** — Implements the standard Tabularis plugin protocol over stdio.
* **Cross-platform** — Can be built for Linux, macOS, and Windows.

## SUPPORTED REDIS DATA TYPES

The plugin currently supports viewing data from the following Redis structures by mapping them to virtual tables:

| Redis Type | Description |
| :--- | :--- |
| **String** | Basic key-value pairs (viewable via the `keys` table) |
| **Hash** | Field-value maps |
| **List** | Ordered collections of strings |
| **Set** | Unordered collections of unique strings |
| **Sorted Set (ZSet)** | Collections of unique strings ordered by score |

*(Note: Data exploration is currently read-only/WIP)*

## INSTALLATION

### AUTOMATIC (VIA TABULARIS)

If your version of Tabularis supports plugin management, the Redis plugin can be installed directly from the application.

### MANUAL INSTALLATION

1. Download the latest release for your platform from the Releases page, or build from source.
2. Extract the archive.
3. Copy the executable (`redis-tabularis-plugin` or `redis-tabularis-plugin.exe` on Windows) and `manifest.json` into the Tabularis plugins directory:

| OS | Plugins Directory |
| :--- | :--- |
| **Linux** | `~/.config/tabularis/plugins/redis/` |
| **macOS** | `~/Library/Application Support/com.debba.tabularis/plugins/redis/` |
| **Windows** | `%APPDATA%\tabularis\plugins\redis\` |

4. Restart Tabularis.

## HOW IT WORKS

The plugin is a standalone Go binary that communicates with Tabularis through JSON-RPC 2.0 over stdio:

1. Tabularis spawns the plugin as a child process.
2. Requests are sent as newline-delimited JSON-RPC messages to the plugin's stdin.
3. Responses are written to stdout in the same format.

This architecture keeps the plugin fully isolated.

## VIRTUAL TABLES

Since Redis is a key-value store, this plugin exposes the following virtual tables to allow SQL-like querying within Tabularis:

- `keys`: Columns `key`, `type`, `ttl`, `value`.
- `hashes`: Columns `key`, `field`, `value`. Query with `SELECT * FROM hashes WHERE key = 'yourkey'`.
- `lists`: Columns `key`, `index`, `value`.
- `sets`: Columns `key`, `value`.
- `zsets`: Columns `key`, `value`, `score`.

## SUPPORTED OPERATIONS

| Method | Description |
| :--- | :--- |
| `test_connection` | Verify database connectivity |
| `get_databases` | List logical databases (0-15) |
| `get_tables` | List virtual tables (keys, hashes, lists, sets, zsets) |
| `get_columns` | Get column metadata for a virtual table |
| `execute_query` | Execute basic queries against virtual tables |

*(Other Tabularis standard methods like `get_schemas`, `get_views`, etc., return empty/unsupported responses as they don't apply directly to Redis.)*

## BUILDING FROM SOURCE

### PREREQUISITES

* Go 1.19+

### BUILD

```bash
go build -o redis-tabularis-plugin main.go
```

The executable will be generated in the current directory.

## DEVELOPMENT

The plugin communicates with Tabularis via JSON-RPC 2.0 over `stdin` and `stdout`. For detailed agent/AI development guidance, see [AGENTS.md](AGENTS.md).

### TESTING THE PLUGIN

You can test the plugin manually by running it and sending JSON-RPC requests via terminal:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"test_connection","params":{"params":{"driver":"redis","host":"localhost","port":6379,"database":"0"}}}' | ./redis-tabularis-plugin
```

**Automated Tests:**

Run unit tests (requires `miniredis`):
```bash
go test -v ./...
```

Run End-to-End tests (requires Docker):
```bash
./run_e2e.sh
```

### TECH STACK

* **Language:** Go 1.19+
* **Driver:** `github.com/go-redis/redis/v8`
* **Protocol:** JSON-RPC 2.0 over stdio

## LICENSE

MIT License

# Redis Tabularis Plugin

A Tabularis driver for Redis databases. This plugin enables [Tabularis](https://github.com/debba/tabularis) to connect to Redis and explore its data in a tabular format.

## Features

- **Key Scanning**: View all keys with their types and TTL.
- **Data Type Support**: (WIP) Explore Hashes, Lists, Sets, and Sorted Sets as virtual tables.
- **JSON-RPC 2.0**: Implements the Tabularis plugin protocol.

## Installation

1. Build the plugin:
   ```bash
   go build -o redis-tabularis-plugin main.go
   ```
2. Copy the `redis-tabularis-plugin` executable and `manifest.json` to your Tabularis plugins directory.
   - On macOS: `~/Library/Application Support/com.debba.tabularis/plugins/redis/`
   - On Linux: `~/.config/tabularis/plugins/redis/`
   - On Windows: `%APPDATA%	abularis\pluginsedis`

## Virtual Tables

Since Redis is a key-value store, this plugin exposes the following virtual tables:

- `keys`: Columns `key`, `type`, `ttl`.
- `hashes`: Columns `key`, `field`, `value`. Query with `SELECT * FROM hashes WHERE key = 'yourkey'`.
- `lists`: Columns `key`, `index`, `value`.
- `sets`: Columns `key`, `member`.
- `zsets`: Columns `key`, `member`, `score`.

## Development

The plugin communicates with Tabularis via JSON-RPC 2.0 over `stdin` and `stdout`.

### Prerequisites

- Go 1.19+
- Redis server (for testing)

### Testing

You can test the plugin manually by running it and sending JSON-RPC requests:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"test_connection","params":{"params":{"driver":"redis","host":"localhost","port":6379,"database":"0"}}}' | ./redis-tabularis-plugin
```

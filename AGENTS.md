# Agent Guidance: Redis Tabularis Plugin

This document provides comprehensive instructions for AI agents on how to interact with, develop, test, and maintain the Redis Tabularis Plugin.

## Table of Contents

- [Project Overview](#project-overview)
- [Architecture & Key Components](#architecture--key-components)
- [File Structure](#file-structure)
- [Development Workflow](#development-workflow)
- [Testing Strategy](#testing-strategy)
- [CI/CD Pipeline](#cicd-pipeline)
- [Common Agent Tasks](#common-agent-tasks)
- [Constraints & Best Practices](#constraints--best-practices)
- [Troubleshooting](#troubleshooting)

## Project Overview

This is a [Tabularis](https://github.com/debba/tabularis) plugin that enables exploring Redis databases as if they were relational tables.

### Key Technologies
- **Language:** Go 1.19+
- **Protocol:** JSON-RPC 2.0 over `stdin` (requests) and `stdout` (responses)
- **Primary Library:** `github.com/go-redis/redis/v8`
- **Testing Library:** `github.com/alicebob/miniredis/v2` (in-memory Redis)
- **Build Tool:** GoReleaser for cross-platform releases

### Repository
- **URL:** `github.com/gzamboni/tabularis-redis-plugin-go`
- **Module Path:** `github.com/gzamboni/tabularis-redis-plugin-go`

## Architecture & Key Components

### 1. JSON-RPC Communication

The plugin is a long-running process that implements a request-response loop:

```
Tabularis → stdin → Plugin → Redis → Plugin → stdout → Tabularis
```

**Key Points:**
- **Main Entry Point:** [`./cmd/tabularis-redis-plugin-go`](./cmd/tabularis-redis-plugin-go) contains the main loop and all request handlers
- **Request/Response Types:** Defined in [`./cmd/tabularis-redis-plugin-go`](./cmd/tabularis-redis-plugin-go) as Go structs
- **Protocol:** JSON-RPC 2.0 with newline-delimited messages
- **Error Handling:** All errors must return valid JSON-RPC error responses

**Request Structure:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "method_name",
  "params": { ... }
}
```

**Response Structure:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": { ... }
}
```

**Error Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32603,
    "message": "Error description"
  }
}
```

### 2. Virtual Tables

Since Redis is a key-value store, it is mapped to virtual tables for SQL-like querying:

| Virtual Table | Schema | Redis Command(s) |
| :--- | :--- | :--- |
| `keys` | (key, type, ttl, value) | `SCAN`, `TYPE`, `TTL`, `GET` |
| `hashes` | (key, field, value) | `HGETALL` |
| `lists` | (key, index, value) | `LRANGE` |
| `sets` | (key, member) | `SMEMBERS` |
| `zsets` | (key, member, score) | `ZRANGE` with scores |

**Important:** When implementing new features or queries, ensure they align with these virtual table schemas.

### 3. Plugin Manifest

[`manifest.json`](manifest.json) contains metadata for Tabularis:

```json
{
  "id": "redis",
  "name": "Redis",
  "version": "0.1.0",
  "description": "Tabularis driver for Redis databases",
  "default_port": 6379,
  "executable": "tabularis-redis-plugin-go",
  "capabilities": {
    "schemas": false,
    "views": false,
    "routines": false,
    "file_based": false
  }
}
```

**When to Update:**
- Version changes (follow semantic versioning)
- New capabilities added
- Default connection parameters changed

### 4. Connection Parameters

The plugin accepts these connection parameters:

```go
type ConnectionParams struct {
    Driver   string  `json:"driver"`   // Always "redis"
    Host     *string `json:"host"`     // Default: "localhost"
    Port     *int    `json:"port"`     // Default: 6379
    Database string  `json:"database"` // Redis DB number (0-15)
    Username *string `json:"username"` // Optional (Redis 6+)
    Password *string `json:"password"` // Optional
    SSLMode  *string `json:"ssl_mode"` // Optional
}
```

## File Structure

```
tabularis-redis-plugin-go/
├── ./cmd/tabularis-redis-plugin-go              # Main plugin logic and JSON-RPC handlers
├── main_test.go         # Unit tests using miniredis
├── manifest.json        # Plugin metadata for Tabularis
├── go.mod               # Go module definition
├── go.sum               # Go dependency checksums
├── run_e2e.sh          # E2E test script (Docker-based)
├── seed_redis.go       # Test data seeding utility
├── .goreleaser.yaml    # GoReleaser configuration
├── .github/
│   └── workflows/
│       ├── ci.yml      # CI workflow (build + test)
│       └── release.yml # Release workflow (cross-compile + publish)
├── README.md           # User-facing documentation
├── AGENTS.md           # This file (AI agent guidance)
└── LICENSE             # MIT License
```

## Development Workflow

### Building

```bash
# Standard build
go build -o tabularis-redis-plugin-go ./cmd/tabularis-redis-plugin-go

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o tabularis-redis-plugin-go-linux-amd64 ./cmd/tabularis-redis-plugin-go
GOOS=darwin GOARCH=arm64 go build -o tabularis-redis-plugin-go-darwin-arm64 ./cmd/tabularis-redis-plugin-go
GOOS=windows GOARCH=amd64 go build -o tabularis-redis-plugin-go-windows-amd64.exe ./cmd/tabularis-redis-plugin-go
```

### Manual Testing

Test the plugin by piping JSON-RPC requests:

```bash
# Test connection
echo '{"jsonrpc":"2.0","id":1,"method":"test_connection","params":{"params":{"driver":"redis","host":"localhost","port":6379,"database":"0"}}}' | ./tabularis-redis-plugin-go

# Get databases
echo '{"jsonrpc":"2.0","id":2,"method":"get_databases","params":{"params":{"driver":"redis","host":"localhost","port":6379}}}' | ./tabularis-redis-plugin-go

# Get tables
echo '{"jsonrpc":"2.0","id":3,"method":"get_tables","params":{"params":{"driver":"redis","host":"localhost","port":6379,"database":"0"}}}' | ./tabularis-redis-plugin-go

# Execute query
echo '{"jsonrpc":"2.0","id":4,"method":"execute_query","params":{"params":{"driver":"redis","host":"localhost","port":6379,"database":"0"},"query":"SELECT * FROM keys"}}' | ./tabularis-redis-plugin-go
```

### Local Development Setup

1. **Install Go 1.19+**
2. **Clone the repository:**
   ```bash
   git clone https://github.com/gzamboni/tabularis-redis-plugin-go.git
   cd tabularis-redis-plugin-go
   ```
3. **Install dependencies:**
   ```bash
   go mod download
   ```
4. **Run tests:**
   ```bash
   go test -v ./...
   ```

### Installing the Plugin Locally

To install the plugin for local development and testing in Tabularis:

1. **Build the plugin:**
   ```bash
   go build -o tabularis-redis-plugin-go ./cmd/tabularis-redis-plugin-go
   ```

2. **Create the plugin directory in Tabularis's data folder:**
   - **Linux:** `~/.local/share/tabularis/plugins/redis/`
   - **macOS:** `~/Library/Application Support/com.debba.tabularis/plugins/redis/`
   - **Windows:** `%APPDATA%\com.debba.tabularis\plugins\redis\
   ```bash
   # Linux
   mkdir -p ~/.local/share/tabul`aris/plugins/redis/
   
   # macOS
   mkdir -p ~/Library/Application\ Support/com.debba.tabularis/plugins/redis/
   
   # Windows (PowerShell)
   New-Item -ItemType Directory -Force -Path "$env:APPDATA\com.debba.tabularis\plugins\redis\"
   ```

3. **Copy the plugin files to the directory:**
   ```bash
   # Linux
   cp tabularis-redis-plugin-go ~/.local/share/tabularis/plugins/redis/
   cp manifest.json ~/.local/share/tabularis/plugins/redis/
   cp README.md ~/.local/share/tabularis/plugins/redis/
   cp LICENSE ~/.local/share/tabularis/plugins/redis/
   
   # macOS
   cp tabularis-redis-plugin-go ~/Library/Application\ Support/com.debba.tabularis/plugins/redis/
   cp manifest.json ~/Library/Application\ Support/com.debba.tabularis/plugins/redis/
   cp README.md ~/Library/Application\ Support/com.debba.tabularis/plugins/redis/
   cp LICENSE ~/Library/Application\ Support/com.debba.tabularis/plugins/redis/
   
   # Windows (PowerShell)
   Copy-Item tabularis-redis-plugin-go.exe "$env:APPDATA\com.debba.tabularis\plugins\redis\"
   Copy-Item manifest.json "$env:APPDATA\com.debba.tabularis\plugins\redis\"
   Copy-Item README.md "$env:APPDATA\com.debba.tabularis\plugins\redis\"
   Copy-Item LICENSE "$env:APPDATA\com.debba.tabularis\plugins\redis\"
   ```

4. **On Linux/macOS, make the executable runnable:**
   ```bash
   # Linux
   chmod +x ~/.local/share/tabularis/plugins/redis/tabularis-redis-plugin-go
   
   # macOS
   chmod +x ~/Library/Application\ Support/com.debba.tabularis/plugins/redis/tabularis-redis-plugin-go
   ```

5. **Restart Tabularis** (or install via Settings to hot-reload without restart)

6. **Verify installation:**
   - Open **Settings → Installed Plugins** — your Redis driver should appear
   - Try creating a new connection using the Redis driver from the connection form

**Quick Install Script (Linux/macOS):**
```bash
#!/bin/bash
# Build the plugin
go build -o tabularis-redis-plugin-go ./cmd/tabularis-redis-plugin-go

# Determine plugin directory based on OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    PLUGIN_DIR="$HOME/Library/Application Support/com.debba.tabularis/plugins/redis"
else
    PLUGIN_DIR="$HOME/.local/share/tabularis/plugins/redis"
fi

# Create directory and copy files
mkdir -p "$PLUGIN_DIR"
cp tabularis-redis-plugin-go manifest.json README.md LICENSE "$PLUGIN_DIR/"
chmod +x "$PLUGIN_DIR/tabularis-redis-plugin-go"

echo "Plugin installed to: $PLUGIN_DIR"
echo "Restart Tabularis to use the plugin."
```

## Testing Strategy

### Unit Tests

**Location:** [`main_test.go`](main_test.go)

**Framework:** Go's built-in `testing` package + `miniredis` (in-memory Redis)

**Run:**
```bash
go test -v ./...
```

**Coverage:**
```bash
go test -cover ./...
```

**Key Test Cases:**
- Connection testing (success and failure scenarios)
- Database listing
- Table listing
- Column metadata retrieval
- Query execution for each virtual table
- Error handling

**Example Test Structure:**
```go
func TestMethodName(t *testing.T) {
    // Setup miniredis
    s := miniredis.RunT(t)
    defer s.Close()
    
    // Seed test data
    s.Set("key1", "value1")
    
    // Create request
    req := Request{...}
    
    // Execute
    handleRequest(req)
    
    // Assert response
    // ...
}
```

### End-to-End Tests

**Location:** [`run_e2e.sh`](run_e2e.sh)

**Requirements:** Docker

**Run:**
```bash
chmod +x run_e2e.sh
./run_e2e.sh
```

**What it does:**
1. Starts a Redis container (`redis:7-alpine`)
2. Seeds test data using [`seed_redis.go`](seed_redis.go)
3. Executes real JSON-RPC requests through the plugin
4. Validates responses
5. Cleans up the container

**Test Data:**
- String keys with various TTLs
- Hash with multiple fields
- List with ordered elements
- Set with unique members
- Sorted set with scores

## CI/CD Pipeline

### CI Workflow

**File:** [`.github/workflows/ci.yml`](.github/workflows/ci.yml)

**Trigger:** Push or PR to `main` branch

**Steps:**
1. Checkout code
2. Setup Go 1.19
3. Build plugin (`go build -v ./...`)
4. Run unit tests (`go test -v ./...`)
5. Run E2E tests (`./run_e2e.sh`)

**Status:** Must pass before merging PRs

### Release Workflow

**File:** [`.github/workflows/release.yml`](.github/workflows/release.yml)

**Trigger:** Tag push matching `v*` (e.g., `v0.1.0`, `v1.2.3`)

**Steps:**
1. Checkout code with full history
2. Setup Go 1.19
3. Run GoReleaser
4. Cross-compile for:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64, arm64)
5. Package each binary with `manifest.json`, `README.md`, `LICENSE`
6. Create ZIP archives
7. Publish GitHub Release with all artifacts

**Creating a Release:**
```bash
# Tag the commit
git tag v0.2.0

# Push the tag
git push origin v0.2.0

# GitHub Actions will automatically build and release
```

**GoReleaser Configuration:** [`.goreleaser.yaml`](.goreleaser.yaml)

## Common Agent Tasks

### Task 1: Adding a New Virtual Table

**Example:** Adding a `streams` table for Redis Streams

1. **Update `get_tables` handler in [`./cmd/tabularis-redis-plugin-go`](./cmd/tabularis-redis-plugin-go):**
   ```go
   tables = append(tables, map[string]interface{}{
       "name": "streams",
       "type": "table",
   })
   ```

2. **Update `get_columns` handler (or `getTableColumns` function):**
   ```go
   case "streams":
       return []map[string]interface{}{
           {"name": "key", "type": "STRING"},
           {"name": "id", "type": "STRING"},
           {"name": "field", "type": "STRING"},
           {"name": "value", "type": "STRING"},
       }
   ```

3. **Implement data fetching in `execute_query` handler:**
   ```go
   case "streams":
       // Use XRANGE or XREAD to fetch stream entries
       // Parse and format as rows
   ```

4. **Add unit tests in [`main_test.go`](main_test.go):**
   ```go
   func TestExecuteQueryStreams(t *testing.T) {
       s := miniredis.RunT(t)
       defer s.Close()
       
       // Seed stream data
       s.XAdd("mystream", "*", []string{"field1", "value1"})
       
       // Test query execution
       // ...
   }
   ```

5. **Update documentation:**
   - Add to virtual tables section in [`README.md`](README.md)
   - Add to virtual tables list in [`AGENTS.md`](AGENTS.md)

### Task 2: Working with Query Support

The plugin implements an **enhanced SQL parser** with comprehensive query support:

**Supported Features:**
- Table extraction from `FROM` clauses
- WHERE conditions with multiple operators (`=`, `!=`, `>`, `<`, `>=`, `<=`, `LIKE`, `IN`)
- Multiple conditions with `AND` operator
- Pattern matching with `LIKE` (supports `%` and `_` wildcards)
- `IN` operator for matching against multiple values
- `ORDER BY` with single or multiple columns (ASC/DESC)
- `LIMIT` and `OFFSET` clauses
- Column filtering on all virtual table columns
- Complex SQL features NOT supported (JOINs, subqueries, aggregations, GROUP BY)

**Implementation Details:**

The query parser is located in [`./cmd/tabularis-redis-plugin-go`](./cmd/tabularis-redis-plugin-go) and consists of:

1. **`parseQuery(query string) QueryParser`** - Main parser function
   - Extracts table name from FROM clause
   - Parses WHERE conditions (supports AND)
   - Extracts LIMIT and OFFSET using regex
   - Returns structured QueryParser object

2. **`parseCondition(condStr string) QueryCondition`** - Condition parser
   - Parses individual conditions like `key = 'value'` or `score > 100`
   - Handles IN operator specially (e.g., `type IN ('hash', 'set')`)
   - Handles operators in order of length (longest first)
   - Strips quotes from values

3. **`matchesConditions(row map[string]interface{}, conditions []QueryCondition) bool`** - Filter function
   - Evaluates if a row matches all conditions
   - Supports numeric and string comparisons
   - Implements LIKE pattern matching with regex
   - Implements IN operator by checking value against comma-separated list

4. **`applyLimitOffset(rows [][]interface{}, limit, offset int) [][]interface{}`** - Result limiting
   - Applies LIMIT and OFFSET to result sets
   - Works independently of pagination

**Query Examples:**

```sql
-- Pattern matching
SELECT * FROM keys WHERE key LIKE 'user:%'

-- Numeric comparison
SELECT * FROM zsets WHERE key = 'leaderboard' AND score > 100

-- IN operator (match multiple values)
SELECT * FROM keys WHERE type IN ('hash', 'set')
SELECT * FROM hashes WHERE field IN ('name', 'email', 'age')

-- Combining IN with other conditions
SELECT * FROM keys WHERE type IN ('hash', 'string') AND key LIKE 'user:%'

-- Column filtering
SELECT * FROM hashes WHERE field = 'email' AND value LIKE '%@example.com'

-- Sorting
SELECT * FROM zsets WHERE key = 'leaderboard' ORDER BY score DESC

-- Multiple column sorting
SELECT * FROM keys ORDER BY type ASC, key DESC

-- Pagination
SELECT * FROM keys LIMIT 10 OFFSET 20

-- Complex filtering with sorting
SELECT * FROM zsets WHERE score >= 50 AND score <= 200 ORDER BY score DESC LIMIT 5
```

**Performance Considerations:**

- Still uses `KEYS *` for key scanning (consider SCAN for production)
- Filtering happens in-memory after fetching from Redis
- LIMIT/OFFSET applied after filtering (not pushed down to Redis)
- For large datasets, consider adding cursor-based pagination

**When Extending Query Support:**

1. **Locate parsing logic** in [`parseQuery()`](./cmd/tabularis-redis-plugin-go:169) and [`parseCondition()`](./cmd/tabularis-redis-plugin-go:236)
2. **Add new operators** to the `operators` slice in `parseCondition()`
3. **Implement operator logic** in [`matchesConditions()`](./cmd/tabularis-redis-plugin-go:270)
4. **Update tests** in [`main_test.go`](main_test.go) with new test cases
5. **Document** in [`README.md`](README.md) and this file

### Task 3: Updating Dependencies

```bash
# Update all dependencies
go get -u ./...

# Update specific dependency
go get -u github.com/go-redis/redis/v8

# Tidy up
go mod tidy

# Verify tests still pass
go test -v ./...
```

### Task 4: Debugging JSON-RPC Issues

1. **Enable stderr logging** (safe, won't corrupt stdout):
   ```go
   fmt.Fprintf(os.Stderr, "DEBUG: Received request: %+v\n", req)
   ```

2. **Test with manual requests:**
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"test_connection","params":{"params":{"driver":"redis","host":"localhost","port":6379,"database":"0"}}}' | ./tabularis-redis-plugin-go 2>debug.log
   ```

3. **Check error responses:**
   - Ensure all errors return valid JSON-RPC error objects
   - Use appropriate error codes (-32700 = parse error, -32603 = internal error)

### Task 5: Adding New Connection Parameters

**Example:** Adding TLS support

1. **Update `ConnectionParams` struct in [`./cmd/tabularis-redis-plugin-go`](./cmd/tabularis-redis-plugin-go):**
   ```go
   type ConnectionParams struct {
       // ... existing fields
       TLSEnabled *bool   `json:"tls_enabled"`
       TLSCert    *string `json:"tls_cert"`
   }
   ```

2. **Update `getClient` function:**
   ```go
   if p.TLSEnabled != nil && *p.TLSEnabled {
       opts.TLSConfig = &tls.Config{...}
   }
   ```

3. **Update [`manifest.json`](manifest.json) if needed:**
   ```json
   {
     "default_tls_enabled": false
   }
   ```

4. **Document in [`README.md`](README.md)**

## Constraints & Best Practices

### Critical Constraints

1. **Standard Output (stdout):**
   - **NEVER** use `fmt.Print`, `fmt.Println`, `log.Print`, or `println` to stdout
   - **NEVER** write debug messages to stdout
   - **ALWAYS** use `fmt.Fprintf(os.Stderr, ...)` for logging
   - **ONLY** write JSON-RPC responses to stdout
   - **Reason:** Tabularis reads stdout as JSON-RPC messages; any extra output corrupts the stream

2. **Error Handling:**
   - **NEVER** panic or crash on errors
   - **ALWAYS** return valid JSON-RPC error responses
   - Use appropriate error codes
   - Provide descriptive error messages

3. **Concurrency:**
   - The plugin handles requests **sequentially** (one at a time)
   - No need for mutex locks or goroutine synchronization
   - Each request is independent

### Best Practices

1. **Redis Performance:**
   - Use `SCAN` instead of `KEYS *` for key iteration
   - Implement cursor-based pagination for large datasets
   - Set reasonable timeouts on Redis operations
   - Consider `LIMIT` clauses to prevent large result sets

2. **Code Organization:**
   - Keep all logic in [`./cmd/tabularis-redis-plugin-go`](./cmd/tabularis-redis-plugin-go) (single-file plugin)
   - Use helper functions for complex operations
   - Document non-obvious logic with comments

3. **Testing:**
   - Write unit tests for every new feature
   - Use `miniredis` for fast, isolated tests
   - Add E2E tests for critical user flows
   - Test error scenarios (connection failures, invalid queries)

4. **Versioning:**
   - Follow semantic versioning (MAJOR.MINOR.PATCH)
   - Update [`manifest.json`](manifest.json) version on releases
   - Create git tags for releases (`v0.1.0`, `v1.0.0`)

5. **Documentation:**
   - Update [`README.md`](README.md) for user-facing changes
   - Update [`AGENTS.md`](AGENTS.md) for architecture changes
   - Add code comments for complex logic
   - Document breaking changes in release notes

## Troubleshooting

### Common Issues

#### Issue: Plugin doesn't respond to requests

**Symptoms:** Tabularis hangs or times out

**Possible Causes:**
1. Plugin crashed (check stderr logs)
2. Invalid JSON-RPC response format
3. Stdout corruption from debug prints

**Solutions:**
1. Check for panics or crashes
2. Validate JSON-RPC response structure
3. Remove any stdout debug prints
4. Test manually with `echo | ./tabularis-redis-plugin-go`

#### Issue: Connection to Redis fails

**Symptoms:** `test_connection` returns error

**Possible Causes:**
1. Redis not running
2. Incorrect host/port
3. Authentication required but not provided
4. Network issues

**Solutions:**
1. Verify Redis is running: `redis-cli ping`
2. Check connection parameters
3. Test with `redis-cli -h <host> -p <port>`
4. Check firewall rules

#### Issue: Query returns no results

**Symptoms:** Empty result set when data exists

**Possible Causes:**
1. Wrong database selected (Redis has 0-15)
2. Query parsing issue
3. Key pattern doesn't match

**Solutions:**
1. Verify database number in connection params
2. Check query parsing logic
3. Test with `redis-cli -n <db> KEYS *`
4. Add debug logging to stderr

#### Issue: Tests fail in CI but pass locally

**Possible Causes:**
1. Docker not available
2. Port conflicts
3. Timing issues

**Solutions:**
1. Check CI logs for specific errors
2. Ensure E2E script handles cleanup
3. Add retries for flaky tests
4. Use unique ports for test Redis instances

### Debugging Checklist

- [ ] Check stderr logs for errors
- [ ] Verify JSON-RPC request format
- [ ] Test Redis connection with `redis-cli`
- [ ] Run unit tests: `go test -v ./...`
- [ ] Run E2E tests: `./run_e2e.sh`
- [ ] Check for stdout corruption
- [ ] Validate response JSON structure
- [ ] Review recent code changes
- [ ] Check dependency versions

---

**For Questions or Issues:**
- Review [`README.md`](README.md) for user documentation
- Check existing tests in [`main_test.go`](main_test.go) for examples
- Examine [`./cmd/tabularis-redis-plugin-go`](./cmd/tabularis-redis-plugin-go) for implementation details
- Consult [Tabularis documentation](https://github.com/debba/tabularis)

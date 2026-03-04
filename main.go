package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var out io.Writer = os.Stdout

type ConnectionParams struct {
	Driver   string  `json:"driver"`
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Database string  `json:"database"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	SSLMode  *string `json:"ssl_mode"`
}

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			sendError(nil, -32700, "Parse error")
			continue
		}
		handleRequest(req)
	}
}

func sendResponse(id json.RawMessage, result interface{}, err error) {
	if err != nil {
		sendError(id, -32603, err.Error())
		return
	}
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintln(out, string(data))
}

func sendError(id json.RawMessage, code int, message string) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: RPCError{
			Code:    code,
			Message: message,
		},
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintln(out, string(data))
}

func getClient(p ConnectionParams) (*redis.Client, error) {
	addr := "localhost:6379"
	if p.Host != nil && p.Port != nil {
		addr = fmt.Sprintf("%s:%d", *p.Host, *p.Port)
	} else if p.Host != nil {
		addr = fmt.Sprintf("%s:6379", *p.Host)
	}

	db, _ := strconv.Atoi(p.Database)
	opts := &redis.Options{
		Addr: addr,
		DB:   db,
	}
	if p.Username != nil && *p.Username != "" {
		if !(*p.Username == "default" && (p.Password == nil || *p.Password == "")) {
			opts.Username = *p.Username
		}
	}
	if p.Password != nil && *p.Password != "" {
		opts.Password = *p.Password
	}

	return redis.NewClient(opts), nil
}

func getTableColumns(table string) []map[string]interface{} {
	switch table {
	case "keys":
		return []map[string]interface{}{
			{"name": "key", "data_type": "STRING", "is_pk": true, "is_nullable": false, "is_auto_increment": false},
			{"name": "type", "data_type": "STRING", "is_pk": false, "is_nullable": false, "is_auto_increment": false},
			{"name": "ttl", "data_type": "INTEGER", "is_pk": false, "is_nullable": false, "is_auto_increment": false},
		}
	case "hashes":
		return []map[string]interface{}{
			{"name": "key", "data_type": "STRING", "is_pk": true, "is_nullable": false, "is_auto_increment": false},
			{"name": "field", "data_type": "STRING", "is_pk": true, "is_nullable": false, "is_auto_increment": false},
			{"name": "value", "data_type": "STRING", "is_pk": false, "is_nullable": false, "is_auto_increment": false},
		}
	case "lists":
		return []map[string]interface{}{
			{"name": "key", "data_type": "STRING", "is_pk": true, "is_nullable": false, "is_auto_increment": false},
			{"name": "index", "data_type": "INTEGER", "is_pk": true, "is_nullable": false, "is_auto_increment": false},
			{"name": "value", "data_type": "STRING", "is_pk": false, "is_nullable": false, "is_auto_increment": false},
		}
	case "sets":
		return []map[string]interface{}{
			{"name": "key", "data_type": "STRING", "is_pk": true, "is_nullable": false, "is_auto_increment": false},
			{"name": "member", "data_type": "STRING", "is_pk": true, "is_nullable": false, "is_auto_increment": false},
		}
	case "zsets":
		return []map[string]interface{}{
			{"name": "key", "data_type": "STRING", "is_pk": true, "is_nullable": false, "is_auto_increment": false},
			{"name": "member", "data_type": "STRING", "is_pk": true, "is_nullable": false, "is_auto_increment": false},
			{"name": "score", "data_type": "NUMERIC", "is_pk": false, "is_nullable": false, "is_auto_increment": false},
		}
	}
	return []map[string]interface{}{}
}

func handleRequest(req Request) {
	switch req.Method {
	case "test_connection":
		var p struct {
			Params ConnectionParams `json:"params"`
		}
		json.Unmarshal(req.Params, &p)
		client, _ := getClient(p.Params)
		err := client.Ping(ctx).Err()
		if err != nil {
			sendResponse(req.ID, map[string]bool{"success": false}, nil)
		} else {
			sendResponse(req.ID, map[string]bool{"success": true}, nil)
		}

	case "get_databases":
		dbs := []string{}
		for i := 0; i < 16; i++ {
			dbs = append(dbs, strconv.Itoa(i))
		}
		sendResponse(req.ID, dbs, nil)

	case "get_schemas", "get_views", "get_routines", "get_routine_parameters", "get_foreign_keys", "get_indexes", "get_all_columns_batch", "get_all_foreign_keys_batch":
		sendResponse(req.ID, []interface{}{}, nil)

	case "get_tables":
		tables := []map[string]interface{}{
			{"name": "keys", "schema": nil, "comment": "All Redis keys"},
			{"name": "hashes", "schema": nil, "comment": "All fields in hashes"},
			{"name": "lists", "schema": nil, "comment": "All elements in lists"},
			{"name": "sets", "schema": nil, "comment": "All members in sets"},
			{"name": "zsets", "schema": nil, "comment": "All members in sorted sets"},
		}
		sendResponse(req.ID, tables, nil)

	case "get_columns":
		var p struct {
			Table string `json:"table"`
		}
		json.Unmarshal(req.Params, &p)
		sendResponse(req.ID, getTableColumns(p.Table), nil)

	case "get_schema_snapshot":
		tables := []string{"keys", "hashes", "lists", "sets", "zsets"}
		snapshotTables := []map[string]interface{}{}
		snapshotColumns := map[string]interface{}{}
		for _, t := range tables {
			snapshotTables = append(snapshotTables, map[string]interface{}{"name": t, "schema": nil})
			snapshotColumns[t] = getTableColumns(t)
		}
		snapshot := map[string]interface{}{
			"tables":       snapshotTables,
			"columns":      snapshotColumns,
			"foreign_keys": map[string]interface{}{},
		}
		sendResponse(req.ID, snapshot, nil)

	case "execute_query":
		var p struct {
			Params   ConnectionParams `json:"params"`
			Query    string           `json:"query"`
			Page     int              `json:"page"`
			PageSize int              `json:"page_size"`
		}
		json.Unmarshal(req.Params, &p)

		query := strings.TrimSpace(p.Query)
		upperQuery := strings.ToUpper(query)
		normalizedQuery := strings.ReplaceAll(upperQuery, "\"", "")
		var result map[string]interface{}

		if strings.Contains(normalizedQuery, "FROM KEYS") {
			result = executeScanKeys(p.Params, p.Page, p.PageSize)
		} else if strings.Contains(normalizedQuery, "FROM HASHES") {
			result = executeScanHashes(p.Params, query, p.Page, p.PageSize)
		} else if strings.Contains(normalizedQuery, "FROM LISTS") {
			result = executeScanLists(p.Params, query, p.Page, p.PageSize)
		} else if strings.Contains(normalizedQuery, "FROM SETS") {
			result = executeScanSets(p.Params, query, p.Page, p.PageSize)
		} else if strings.Contains(normalizedQuery, "FROM ZSETS") {
			result = executeScanZSets(p.Params, query, p.Page, p.PageSize)
		} else {
			result = map[string]interface{}{
				"columns":       []string{"error"},
				"rows":          [][]interface{}{{fmt.Sprintf("Unsupported query: %s. Try 'SELECT * FROM keys' or 'SELECT * FROM hashes WHERE key = \"mykey\"'", p.Query)}},
				"affected_rows": 0,
				"truncated":     false,
				"has_more":      false,
				"pagination":    nil,
			}
		}

		// result already has pagination info from executors
		sendResponse(req.ID, result, nil)

	default:
		sendError(req.ID, -32601, "Method not found")
	}
}

func executeScanKeys(p ConnectionParams, page, pageSize int) map[string]interface{} {
	client, _ := getClient(p)
	keys, _ := client.Keys(ctx, "*").Result()
	
	total := len(keys)
	if pageSize == 0 { pageSize = 50 }
	if page == 0 { page = 1 }
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize
	if startIdx > total { startIdx = total }
	if endIdx > total { endIdx = total }
	
	pagedKeys := keys[startIdx:endIdx]
	rows := [][]interface{}{}
	for _, k := range pagedKeys {
		typ, _ := client.Type(ctx, k).Result()
		ttl, _ := client.TTL(ctx, k).Result()
		rows = append(rows, []interface{}{k, typ, int64(ttl.Seconds())})
	}
	
	hasMore := endIdx < total
	return map[string]interface{}{
		"columns":       []string{"key", "type", "ttl"},
		"rows":          rows,
		"affected_rows": 0,
		"truncated":     hasMore,
		"pagination": map[string]interface{}{
			"page":       page,
			"page_size":  pageSize,
			"total_rows": total,
			"has_more":   hasMore,
		},
	}
}

func extractKey(query string) string {
	upperQuery := strings.ToUpper(query)
	idx := strings.Index(upperQuery, "WHERE KEY =")
	if idx != -1 {
		valPart := query[idx+len("WHERE KEY ="):]
		spaceIdx := strings.Index(strings.TrimSpace(valPart), " ")
		if spaceIdx != -1 {
			valPart = strings.TrimSpace(valPart)[:spaceIdx]
		}
		return strings.TrimSpace(strings.Trim(strings.TrimSpace(valPart), " '\""))
	}
	return ""
}

func paginateRows(rows [][]interface{}, page, pageSize int) ([][]interface{}, bool, int) {
	total := len(rows)
	if pageSize == 0 { pageSize = 50 }
	if page == 0 { page = 1 }
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize
	if startIdx > total { startIdx = total }
	if endIdx > total { endIdx = total }
	
	hasMore := endIdx < total
	return rows[startIdx:endIdx], hasMore, total
}

func executeScanHashes(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
	client, _ := getClient(p)
	key := extractKey(query)

	rows := [][]interface{}{}

	if key != "" {
		fields, _ := client.HGetAll(ctx, key).Result()
		for f, v := range fields {
			rows = append(rows, []interface{}{key, f, v})
		}
	} else {
		keys, _ := client.Keys(ctx, "*").Result()
		for _, k := range keys {
			typ, _ := client.Type(ctx, k).Result()
			if typ == "hash" {
				fields, _ := client.HGetAll(ctx, k).Result()
				for f, v := range fields {
					rows = append(rows, []interface{}{k, f, v})
				}
			}
		}
	}
	
	pagedRows, hasMore, total := paginateRows(rows, page, pageSize)
	
	return map[string]interface{}{
		"columns": []string{"key", "field", "value"},
		"rows": pagedRows,
		"affected_rows": 0,
		"truncated": hasMore,
		"pagination": map[string]interface{}{
			"page":       page,
			"page_size":  pageSize,
			"total_rows": total,
			"has_more":   hasMore,
		},
	}
}

func executeScanLists(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
	client, _ := getClient(p)
	key := extractKey(query)

	rows := [][]interface{}{}

	if key != "" {
		values, _ := client.LRange(ctx, key, 0, -1).Result()
		for i, v := range values {
			rows = append(rows, []interface{}{key, i, v})
		}
	} else {
		keys, _ := client.Keys(ctx, "*").Result()
		for _, k := range keys {
			typ, _ := client.Type(ctx, k).Result()
			if typ == "list" {
				values, _ := client.LRange(ctx, k, 0, -1).Result()
				for i, v := range values {
					rows = append(rows, []interface{}{k, i, v})
				}
			}
		}
	}
	
	pagedRows, hasMore, total := paginateRows(rows, page, pageSize)

	return map[string]interface{}{
		"columns": []string{"key", "index", "value"},
		"rows": pagedRows,
		"affected_rows": 0,
		"truncated": hasMore,
		"pagination": map[string]interface{}{
			"page":       page,
			"page_size":  pageSize,
			"total_rows": total,
			"has_more":   hasMore,
		},
	}
}

func executeScanSets(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
	client, _ := getClient(p)
	key := extractKey(query)

	rows := [][]interface{}{}

	if key != "" {
		members, _ := client.SMembers(ctx, key).Result()
		for _, m := range members {
			rows = append(rows, []interface{}{key, m})
		}
	} else {
		keys, _ := client.Keys(ctx, "*").Result()
		for _, k := range keys {
			typ, _ := client.Type(ctx, k).Result()
			if typ == "set" {
				members, _ := client.SMembers(ctx, k).Result()
				for _, m := range members {
					rows = append(rows, []interface{}{k, m})
				}
			}
		}
	}
	
	pagedRows, hasMore, total := paginateRows(rows, page, pageSize)

	return map[string]interface{}{
		"columns": []string{"key", "member"},
		"rows": pagedRows,
		"affected_rows": 0,
		"truncated": hasMore,
		"pagination": map[string]interface{}{
			"page":       page,
			"page_size":  pageSize,
			"total_rows": total,
			"has_more":   hasMore,
		},
	}
}

func executeScanZSets(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
	client, _ := getClient(p)
	key := extractKey(query)

	rows := [][]interface{}{}

	if key != "" {
		members, _ := client.ZRangeWithScores(ctx, key, 0, -1).Result()
		for _, m := range members {
			rows = append(rows, []interface{}{key, m.Member, m.Score})
		}
	} else {
		keys, _ := client.Keys(ctx, "*").Result()
		for _, k := range keys {
			typ, _ := client.Type(ctx, k).Result()
			if typ == "zset" {
				members, _ := client.ZRangeWithScores(ctx, k, 0, -1).Result()
				for _, m := range members {
					rows = append(rows, []interface{}{k, m.Member, m.Score})
				}
			}
		}
	}
	
	pagedRows, hasMore, total := paginateRows(rows, page, pageSize)

	return map[string]interface{}{
		"columns": []string{"key", "member", "score"},
		"rows": pagedRows,
		"affected_rows": 0,
		"truncated": hasMore,
		"pagination": map[string]interface{}{
			"page":       page,
			"page_size":  pageSize,
			"total_rows": total,
			"has_more":   hasMore,
		},
	}
}

import re

with open('internal/plugin/executor.go', 'r') as f:
    content = f.read()

# Hashes
old_hash = """func executeScanHashes(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
	client, _ := getClient(p)
	key := extractKey(query)

	if key == "" {
		return map[string]interface{}{
			"columns": []string{"key", "field", "value"},
			"rows": [][]interface{}{},
			"affected_rows": 0,
			"truncated": false,
			"pagination": nil,
		}
	}

	fields, _ := client.HGetAll(ctx, key).Result()
	rows := [][]interface{}{}
	for f, v := range fields {
		rows = append(rows, []interface{}{key, f, v})
	}"""
new_hash = """func executeScanHashes(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
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
	}"""
content = content.replace(old_hash, new_hash)

# Lists
old_list = """func executeScanLists(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
	client, _ := getClient(p)
	key := extractKey(query)

	if key == "" {
		return map[string]interface{}{
			"columns": []string{"key", "index", "value"},
			"rows": [][]interface{}{},
			"affected_rows": 0,
			"truncated": false,
			"pagination": nil,
		}
	}

	values, _ := client.LRange(ctx, key, 0, -1).Result()
	rows := [][]interface{}{}
	for i, v := range values {
		rows = append(rows, []interface{}{key, i, v})
	}"""
new_list = """func executeScanLists(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
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
	}"""
content = content.replace(old_list, new_list)

# Sets
old_set = """func executeScanSets(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
	client, _ := getClient(p)
	key := extractKey(query)

	if key == "" {
		return map[string]interface{}{
			"columns": []string{"key", "member"},
			"rows": [][]interface{}{},
			"affected_rows": 0,
			"truncated": false,
			"pagination": nil,
		}
	}

	members, _ := client.SMembers(ctx, key).Result()
	rows := [][]interface{}{}
	for _, m := range members {
		rows = append(rows, []interface{}{key, m})
	}"""
new_set = """func executeScanSets(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
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
	}"""
content = content.replace(old_set, new_set)

# ZSets
old_zset = """func executeScanZSets(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
	client, _ := getClient(p)
	key := extractKey(query)

	if key == "" {
		return map[string]interface{}{
			"columns": []string{"key", "member", "score"},
			"rows": [][]interface{}{},
			"affected_rows": 0,
			"truncated": false,
			"pagination": nil,
		}
	}

	members, _ := client.ZRangeWithScores(ctx, key, 0, -1).Result()
	rows := [][]interface{}{}
	for _, m := range members {
		rows = append(rows, []interface{}{key, m.Member, m.Score})
	}"""
new_zset = """func executeScanZSets(p ConnectionParams, query string, page, pageSize int) map[string]interface{} {
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
	}"""
content = content.replace(old_zset, new_zset)

with open('internal/plugin/executor.go', 'w') as f:
    f.write(content)

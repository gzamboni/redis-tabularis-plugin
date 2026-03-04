#!/bin/bash
set -e

echo "Building plugin..."
go build -o redis-tabularis-plugin main.go

echo "Starting Redis via Docker..."
# Use port 6380 to avoid conflicts with any local Redis instance
CONTAINER_ID=$(docker run -d -p 6380:6379 redis:alpine)

# Ensure container is destroyed when script exits
trap "echo 'Cleaning up...'; docker rm -f $CONTAINER_ID > /dev/null" EXIT

echo "Waiting for Redis to start..."
sleep 2

echo "Seeding data..."
docker exec $CONTAINER_ID redis-cli set e2e_key "hello" > /dev/null
docker exec $CONTAINER_ID redis-cli hset e2e_hash myfield "myvalue" > /dev/null

echo "Running E2E tests..."

echo "Test 1: test_connection"
RESULT_1=$(echo '{"jsonrpc":"2.0","id":1,"method":"test_connection","params":{"params":{"driver":"redis","host":"localhost","port":6380,"database":"0"}}}' | ./redis-tabularis-plugin)
if echo "$RESULT_1" | grep -q '"success":true'; then
    echo "✅ test_connection passed"
else
    echo "❌ test_connection failed: $RESULT_1"
    exit 1
fi

echo "Test 2: execute_query (keys)"
RESULT_2=$(echo '{"jsonrpc":"2.0","id":2,"method":"execute_query","params":{"params":{"driver":"redis","host":"localhost","port":6380,"database":"0"}, "query":"SELECT * FROM keys", "page":0, "page_size":10}}' | ./redis-tabularis-plugin)
if echo "$RESULT_2" | grep -q '"e2e_key"'; then
    echo "✅ execute_query (keys) passed"
else
    echo "❌ execute_query (keys) failed: $RESULT_2"
    exit 1
fi

echo "Test 3: execute_query (hashes)"
RESULT_3=$(echo '{"jsonrpc":"2.0","id":3,"method":"execute_query","params":{"params":{"driver":"redis","host":"localhost","port":6380,"database":"0"}, "query":"SELECT * FROM hashes WHERE key = '"'e2e_hash'"'", "page":0, "page_size":10}}' | ./redis-tabularis-plugin)
if echo "$RESULT_3" | grep -q '"myvalue"'; then
    echo "✅ execute_query (hashes) passed"
else
    echo "❌ execute_query (hashes) failed: $RESULT_3"
    exit 1
fi

echo "🎉 All E2E tests passed successfully!"

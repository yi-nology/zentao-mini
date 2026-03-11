#!/bin/bash

# 测试MCP服务的ping功能
echo "Testing MCP ping..."
echo '{"action": "ping"}' | go run main.go 2>&1 | grep -A 3 'status'

echo "\nTesting MCP get_products..."
echo '{"action": "get_products"}' | go run main.go 2>&1 | grep -A 5 'status'

echo "\nTesting MCP get_projects..."
echo '{"action": "get_projects"}' | go run main.go 2>&1 | grep -A 5 'status'

echo "\nTesting MCP get_bugs..."
echo '{"action": "get_bugs"}' | go run main.go 2>&1 | grep -A 5 'status'

echo "\nTesting MCP get_stories..."
echo '{"action": "get_stories"}' | go run main.go 2>&1 | grep -A 5 'status'

echo "\nTesting MCP get_tasks..."
echo '{"action": "get_tasks"}' | go run main.go 2>&1 | grep -A 5 'status'

echo "\nTesting MCP get_users..."
echo '{"action": "get_users"}' | go run main.go 2>&1 | grep -A 5 'status'

echo "\nTesting MCP get_timelog..."
echo '{"action": "get_timelog"}' | go run main.go 2>&1 | grep -A 5 'status'
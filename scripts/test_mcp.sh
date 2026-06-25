#!/bin/bash

# 测试MCP服务的HTTP端点
# 用法: ./scripts/test_mcp.sh [base_url]

BASE_URL="${1:-http://localhost:12345}"

echo "Testing MCP endpoints at ${BASE_URL}..."
echo ""

echo "=== Testing MCP ping ==="
curl -s -X POST "${BASE_URL}/mcp/ping" -H "Content-Type: application/json" | head -c 200
echo ""

echo "=== Testing MCP get_products ==="
curl -s -X POST "${BASE_URL}/mcp/products" -H "Content-Type: application/json" | head -c 200
echo ""

echo "=== Testing MCP get_projects ==="
curl -s -X POST "${BASE_URL}/mcp/projects" -H "Content-Type: application/json" | head -c 200
echo ""

echo "=== Testing MCP get_bugs ==="
curl -s -X POST "${BASE_URL}/mcp/bugs" -H "Content-Type: application/json" | head -c 200
echo ""

echo "=== Testing MCP get_stories ==="
curl -s -X POST "${BASE_URL}/mcp/stories" -H "Content-Type: application/json" | head -c 200
echo ""

echo "=== Testing MCP get_tasks ==="
curl -s -X POST "${BASE_URL}/mcp/tasks" -H "Content-Type: application/json" | head -c 200
echo ""

echo "=== Testing health endpoint ==="
curl -s "${BASE_URL}/health" | head -c 200
echo ""

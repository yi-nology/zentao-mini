<template>
  <div class="mcp-guide-container">
    <h1>MCP 服务对接指南</h1>
    
    <div class="section">
      <h2>支持能力清单</h2>
      <div class="capability-list">
        <div class="capability-item">
          <span class="capability-name">ping</span>
          <span class="capability-desc">测试服务状态</span>
        </div>
        <div class="capability-item">
          <span class="capability-name">get_products</span>
          <span class="capability-desc">获取产品列表</span>
        </div>
        <div class="capability-item">
          <span class="capability-name">get_projects</span>
          <span class="capability-desc">获取项目列表</span>
        </div>
        <div class="capability-item">
          <span class="capability-name">get_executions</span>
          <span class="capability-desc">获取执行/迭代列表</span>
        </div>
        <div class="capability-item">
          <span class="capability-name">get_bugs</span>
          <span class="capability-desc">获取 Bug 列表</span>
        </div>
        <div class="capability-item">
          <span class="capability-name">get_stories</span>
          <span class="capability-desc">获取需求列表</span>
        </div>
        <div class="capability-item">
          <span class="capability-name">get_tasks</span>
          <span class="capability-desc">获取任务列表</span>
        </div>
        <div class="capability-item">
          <span class="capability-name">get_users</span>
          <span class="capability-desc">获取用户列表</span>
        </div>
        <div class="capability-item">
          <span class="capability-name">get_timelog</span>
          <span class="capability-desc">获取工时数据</span>
        </div>
      </div>
    </div>

    <div class="section">
      <h2>软件对接方案</h2>
      
      <div class="software-card">
        <h3>Python 对接</h3>
        <pre><code>import json
import subprocess

# 启动应用
process = subprocess.Popen(
    ["./chandao-mini"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    text=True
)

# 发送 ping 请求
request = {"action": "ping"}
process.stdin.write(json.dumps(request) + "\n")
process.stdin.flush()

# 接收响应
response = process.stdout.readline()
print("Response:", response)

# 关闭进程
process.stdin.close()
process.terminate()</code></pre>
      </div>

      <div class="software-card">
        <h3>Bash 对接</h3>
        <pre><code>#!/bin/bash

# 测试 ping 动作
echo '{"action": "ping"}' | ./chandao-mini 2>&1 | grep -A 3 'status'

# 测试获取产品列表
echo '{"action": "get_products"}' | ./chandao-mini 2>&1 | grep -A 5 'status'</code></pre>
      </div>

      <div class="software-card">
        <h3>Node.js 对接</h3>
        <pre><code>const { spawn } = require('child_process');

// 启动应用
const process = spawn('./chandao-mini');

// 发送 ping 请求
const request = { action: 'ping' };
process.stdin.write(JSON.stringify(request) + '\n');
process.stdin.end();

// 接收响应
process.stdout.on('data', (data) => {
  console.log('Response:', data.toString());
});

// 处理错误
process.stderr.on('data', (data) => {
  console.error('Error:', data.toString());
});</code></pre>
      </div>

      <div class="software-card">
        <h3>Go 对接</h3>
        <pre><code>package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func main() {
	// 启动应用
	cmd := exec.Command("./chandao-mini")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	// 发送 ping 请求
	request := map[string]interface{}{
		"action": "ping",
	}
	requestJSON, _ := json.Marshal(request)
	stdin.Write(append(requestJSON, '\n'))
	stdin.Close()

	// 接收响应
	responseJSON := make([]byte, 1024)
	n, _ := stdout.Read(responseJSON)
	fmt.Println("Response:", string(responseJSON[:n]))

	cmd.Wait()
}</code></pre>
      </div>

      <div class="software-card">
        <h3>Qoder 对接</h3>
        <pre><code>// Qoder 插件示例
const { spawn } = require('child_process');

class ChandaoMiniPlugin {
  constructor() {
    this.process = null;
  }

  async start() {
    // 启动禅道迷你应用
    this.process = spawn('./chandao-mini');
    return 'Chandao Mini MCP service started';
  }

  async call(action, params = {}) {
    if (!this.process) {
      await this.start();
    }

    return new Promise((resolve, reject) => {
      // 发送请求
      const request = { action, params };
      this.process.stdin.write(JSON.stringify(request) + '\n');

      // 接收响应
      const handleResponse = (data) => {
        const response = JSON.parse(data.toString());
        resolve(response);
        // 移除监听器
        this.process.stdout.off('data', handleResponse);
      };

      this.process.stdout.on('data', handleResponse);

      // 处理错误
      this.process.stderr.on('data', (data) => {
        reject(new Error(data.toString()));
      });
    });
  }

  async stop() {
    if (this.process) {
      this.process.stdin.end();
      this.process.terminate();
      this.process = null;
      return 'Chandao Mini MCP service stopped';
    }
    return 'Service not running';
  }
}

// 使用示例
const plugin = new ChandaoMiniPlugin();
plugin.start().then(() => {
  plugin.call('ping').then(response => {
    console.log('Ping response:', response);
  });
});</code></pre>
      </div>

      <div class="software-card">
        <h3>Trae Skill 对接</h3>
        <pre><code>// Trae Skill 示例
module.exports = {
  name: 'chandao-mini',
  version: '1.0.0',
  description: '禅道迷你应用 MCP 服务对接',
  
  actions: {
    async ping(ctx) {
      return this.callMCP('ping');
    },
    
    async getProducts(ctx) {
      return this.callMCP('get_products');
    },
    
    async getBugs(ctx) {
      return this.callMCP('get_bugs', ctx.params);
    }
  },
  
  methods: {
    callMCP(action, params = {}) {
      const { spawnSync } = require('child_process');
      
      // 构建请求
      const request = JSON.stringify({ action, params });
      
      // 执行命令
      const result = spawnSync('./chandao-mini', {
        input: request + '\n',
        encoding: 'utf8'
      });
      
      // 解析响应
      if (result.stdout) {
        try {
          return JSON.parse(result.stdout);
        } catch (error) {
          return { status: 'error', message: 'Invalid response format' };
        }
      } else {
        return { status: 'error', message: result.stderr || 'No response' };
      }
    }
  }
};</code></pre>
      </div>
    </div>

    <div class="section">
      <h2>请求格式</h2>
      <pre><code>{"action": "动作名称", "params": {"参数1": "值1", "参数2": "值2"}}</code></pre>
    </div>

    <div class="section">
      <h2>响应格式</h2>
      <pre><code>// 成功响应
{"status": "ok", "message": "操作成功", "data": [...]}

// 错误响应
{"status": "error", "message": "错误信息"}</code></pre>
    </div>
  </div>
</template>

<style scoped>
.mcp-guide-container {
  padding: 20px;
  max-width: 1000px;
  margin: 0 auto;
}

h1 {
  color: #333;
  margin-bottom: 40px;
  text-align: center;
  font-size: 28px;
  font-weight: 600;
}

h2 {
  color: #4CAF50;
  margin: 30px 0 20px;
  font-size: 20px;
  font-weight: 500;
  border-bottom: 1px solid #e0e0e0;
  padding-bottom: 10px;
}

h3 {
  color: #555;
  margin: 20px 0 15px;
  font-size: 16px;
  font-weight: 500;
}

.section {
  background: #ffffff;
  border-radius: 8px;
  padding: 25px;
  margin-bottom: 30px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.capability-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 15px;
}

.capability-item {
  background: #f8f9fa;
  padding: 15px;
  border-radius: 6px;
  border-left: 4px solid #4CAF50;
  transition: transform 0.2s, box-shadow 0.2s;
}

.capability-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.capability-name {
  display: block;
  font-weight: 600;
  color: #333;
  margin-bottom: 5px;
}

.capability-desc {
  color: #666;
  font-size: 14px;
}

.software-card {
  background: #f8f9fa;
  padding: 20px;
  border-radius: 6px;
  margin-bottom: 20px;
  border-left: 4px solid #2196F3;
}

pre {
  background: #f5f5f5;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 10px 0;
  font-family: 'Courier New', Courier, monospace;
  font-size: 14px;
  line-height: 1.4;
  border: 1px solid #e0e0e0;
}

code {
  color: #333;
}

@media (max-width: 768px) {
  .capability-list {
    grid-template-columns: 1fr;
  }
  
  .mcp-guide-container {
    padding: 10px;
  }
  
  .section {
    padding: 20px;
  }
}
</style>
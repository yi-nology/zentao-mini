import{_ as a,b as n,d as s,J as e}from"./index-759f9a80.js";const d={},o={class:"mcp-guide-container"};function i(c,t){return n(),s("div",o,[...t[0]||(t[0]=[e(`<h1 data-v-ebd4845d>MCP 服务对接指南</h1><div class="section" data-v-ebd4845d><h2 data-v-ebd4845d>支持能力清单</h2><div class="capability-list" data-v-ebd4845d><div class="capability-item" data-v-ebd4845d><span class="capability-name" data-v-ebd4845d>ping</span><span class="capability-desc" data-v-ebd4845d>测试服务状态</span></div><div class="capability-item" data-v-ebd4845d><span class="capability-name" data-v-ebd4845d>get_products</span><span class="capability-desc" data-v-ebd4845d>获取产品列表</span></div><div class="capability-item" data-v-ebd4845d><span class="capability-name" data-v-ebd4845d>get_projects</span><span class="capability-desc" data-v-ebd4845d>获取项目列表</span></div><div class="capability-item" data-v-ebd4845d><span class="capability-name" data-v-ebd4845d>get_executions</span><span class="capability-desc" data-v-ebd4845d>获取执行/迭代列表</span></div><div class="capability-item" data-v-ebd4845d><span class="capability-name" data-v-ebd4845d>get_bugs</span><span class="capability-desc" data-v-ebd4845d>获取 Bug 列表</span></div><div class="capability-item" data-v-ebd4845d><span class="capability-name" data-v-ebd4845d>get_stories</span><span class="capability-desc" data-v-ebd4845d>获取需求列表</span></div><div class="capability-item" data-v-ebd4845d><span class="capability-name" data-v-ebd4845d>get_tasks</span><span class="capability-desc" data-v-ebd4845d>获取任务列表</span></div><div class="capability-item" data-v-ebd4845d><span class="capability-name" data-v-ebd4845d>get_users</span><span class="capability-desc" data-v-ebd4845d>获取用户列表</span></div><div class="capability-item" data-v-ebd4845d><span class="capability-name" data-v-ebd4845d>get_timelog</span><span class="capability-desc" data-v-ebd4845d>获取工时数据</span></div></div></div><div class="section" data-v-ebd4845d><h2 data-v-ebd4845d>软件对接方案</h2><div class="software-card" data-v-ebd4845d><h3 data-v-ebd4845d>Python 对接</h3><pre data-v-ebd4845d><code data-v-ebd4845d>import json
import subprocess

# 启动应用
process = subprocess.Popen(
    [&quot;./chandao-mini&quot;],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    text=True
)

# 发送 ping 请求
request = {&quot;action&quot;: &quot;ping&quot;}
process.stdin.write(json.dumps(request) + &quot;\\n&quot;)
process.stdin.flush()

# 接收响应
response = process.stdout.readline()
print(&quot;Response:&quot;, response)

# 关闭进程
process.stdin.close()
process.terminate()</code></pre></div><div class="software-card" data-v-ebd4845d><h3 data-v-ebd4845d>Bash 对接</h3><pre data-v-ebd4845d><code data-v-ebd4845d>#!/bin/bash

# 测试 ping 动作
echo &#39;{&quot;action&quot;: &quot;ping&quot;}&#39; | ./chandao-mini 2&gt;&amp;1 | grep -A 3 &#39;status&#39;

# 测试获取产品列表
echo &#39;{&quot;action&quot;: &quot;get_products&quot;}&#39; | ./chandao-mini 2&gt;&amp;1 | grep -A 5 &#39;status&#39;</code></pre></div><div class="software-card" data-v-ebd4845d><h3 data-v-ebd4845d>Node.js 对接</h3><pre data-v-ebd4845d><code data-v-ebd4845d>const { spawn } = require(&#39;child_process&#39;);

// 启动应用
const process = spawn(&#39;./chandao-mini&#39;);

// 发送 ping 请求
const request = { action: &#39;ping&#39; };
process.stdin.write(JSON.stringify(request) + &#39;\\n&#39;);
process.stdin.end();

// 接收响应
process.stdout.on(&#39;data&#39;, (data) =&gt; {
  console.log(&#39;Response:&#39;, data.toString());
});

// 处理错误
process.stderr.on(&#39;data&#39;, (data) =&gt; {
  console.error(&#39;Error:&#39;, data.toString());
});</code></pre></div><div class="software-card" data-v-ebd4845d><h3 data-v-ebd4845d>Go 对接</h3><pre data-v-ebd4845d><code data-v-ebd4845d>package main

import (
	&quot;encoding/json&quot;
	&quot;fmt&quot;
	&quot;os/exec&quot;
)

func main() {
	// 启动应用
	cmd := exec.Command(&quot;./chandao-mini&quot;)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	// 发送 ping 请求
	request := map[string]interface{}{
		&quot;action&quot;: &quot;ping&quot;,
	}
	requestJSON, _ := json.Marshal(request)
	stdin.Write(append(requestJSON, &#39;\\n&#39;))
	stdin.Close()

	// 接收响应
	responseJSON := make([]byte, 1024)
	n, _ := stdout.Read(responseJSON)
	fmt.Println(&quot;Response:&quot;, string(responseJSON[:n]))

	cmd.Wait()
}</code></pre></div><div class="software-card" data-v-ebd4845d><h3 data-v-ebd4845d>Qoder 对接</h3><pre data-v-ebd4845d><code data-v-ebd4845d>// Qoder 插件示例
const { spawn } = require(&#39;child_process&#39;);

class ChandaoMiniPlugin {
  constructor() {
    this.process = null;
  }

  async start() {
    // 启动禅道迷你应用
    this.process = spawn(&#39;./chandao-mini&#39;);
    return &#39;Chandao Mini MCP service started&#39;;
  }

  async call(action, params = {}) {
    if (!this.process) {
      await this.start();
    }

    return new Promise((resolve, reject) =&gt; {
      // 发送请求
      const request = { action, params };
      this.process.stdin.write(JSON.stringify(request) + &#39;\\n&#39;);

      // 接收响应
      const handleResponse = (data) =&gt; {
        const response = JSON.parse(data.toString());
        resolve(response);
        // 移除监听器
        this.process.stdout.off(&#39;data&#39;, handleResponse);
      };

      this.process.stdout.on(&#39;data&#39;, handleResponse);

      // 处理错误
      this.process.stderr.on(&#39;data&#39;, (data) =&gt; {
        reject(new Error(data.toString()));
      });
    });
  }

  async stop() {
    if (this.process) {
      this.process.stdin.end();
      this.process.terminate();
      this.process = null;
      return &#39;Chandao Mini MCP service stopped&#39;;
    }
    return &#39;Service not running&#39;;
  }
}

// 使用示例
const plugin = new ChandaoMiniPlugin();
plugin.start().then(() =&gt; {
  plugin.call(&#39;ping&#39;).then(response =&gt; {
    console.log(&#39;Ping response:&#39;, response);
  });
});</code></pre></div><div class="software-card" data-v-ebd4845d><h3 data-v-ebd4845d>Trae Skill 对接</h3><pre data-v-ebd4845d><code data-v-ebd4845d>// Trae Skill 示例
module.exports = {
  name: &#39;chandao-mini&#39;,
  version: &#39;1.0.0&#39;,
  description: &#39;禅道迷你应用 MCP 服务对接&#39;,
  
  actions: {
    async ping(ctx) {
      return this.callMCP(&#39;ping&#39;);
    },
    
    async getProducts(ctx) {
      return this.callMCP(&#39;get_products&#39;);
    },
    
    async getBugs(ctx) {
      return this.callMCP(&#39;get_bugs&#39;, ctx.params);
    }
  },
  
  methods: {
    callMCP(action, params = {}) {
      const { spawnSync } = require(&#39;child_process&#39;);
      
      // 构建请求
      const request = JSON.stringify({ action, params });
      
      // 执行命令
      const result = spawnSync(&#39;./chandao-mini&#39;, {
        input: request + &#39;\\n&#39;,
        encoding: &#39;utf8&#39;
      });
      
      // 解析响应
      if (result.stdout) {
        try {
          return JSON.parse(result.stdout);
        } catch (error) {
          return { status: &#39;error&#39;, message: &#39;Invalid response format&#39; };
        }
      } else {
        return { status: &#39;error&#39;, message: result.stderr || &#39;No response&#39; };
      }
    }
  }
};</code></pre></div></div><div class="section" data-v-ebd4845d><h2 data-v-ebd4845d>请求格式</h2><pre data-v-ebd4845d><code data-v-ebd4845d>{&quot;action&quot;: &quot;动作名称&quot;, &quot;params&quot;: {&quot;参数1&quot;: &quot;值1&quot;, &quot;参数2&quot;: &quot;值2&quot;}}</code></pre></div><div class="section" data-v-ebd4845d><h2 data-v-ebd4845d>响应格式</h2><pre data-v-ebd4845d><code data-v-ebd4845d>// 成功响应
{&quot;status&quot;: &quot;ok&quot;, &quot;message&quot;: &quot;操作成功&quot;, &quot;data&quot;: [...]}

// 错误响应
{&quot;status&quot;: &quot;error&quot;, &quot;message&quot;: &quot;错误信息&quot;}</code></pre></div>`,5)])])}const p=a(d,[["render",i],["__scopeId","data-v-ebd4845d"]]);export{p as default};

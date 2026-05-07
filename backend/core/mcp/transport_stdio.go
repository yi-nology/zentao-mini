package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// StdioTransport stdio 传输层
type StdioTransport struct {
	server *MCPServer
	stdin  io.Reader
	stdout io.Writer
	mutex  sync.Mutex
}

// NewStdioTransport 创建 stdio 传输层
func NewStdioTransport(server *MCPServer) *StdioTransport {
	return &StdioTransport{
		server: server,
		stdin:  os.Stdin,
		stdout: os.Stdout,
	}
}

// NewStdioTransportWithIO 创建带自定义 IO 的 stdio 传输层（用于测试）
func NewStdioTransportWithIO(server *MCPServer, stdin io.Reader, stdout io.Writer) *StdioTransport {
	return &StdioTransport{
		server: server,
		stdin:  stdin,
		stdout: stdout,
	}
}

// Start 启动 stdio 监听
func (t *StdioTransport) Start() {
	go t.listen()
}

func (t *StdioTransport) listen() {
	decoder := json.NewDecoder(t.stdin)
	encoder := json.NewEncoder(t.stdout)

	for {
		var request map[string]interface{}
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				break
			}
			t.sendError(encoder, fmt.Sprintf("Invalid request: %v", err))
			continue
		}

		action, _ := request["action"].(string)
		params, _ := request["params"].(map[string]interface{})

		result, err := t.server.HandleAction(action, params)
		if err != nil {
			t.sendError(encoder, err.Error())
			continue
		}

		t.send(encoder, result)
	}
}

func (t *StdioTransport) send(encoder *json.Encoder, data interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if err := encoder.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send response: %v\n", err)
	}
}

func (t *StdioTransport) sendError(encoder *json.Encoder, msg string) {
	t.send(encoder, map[string]interface{}{
		"status":  "error",
		"message": msg,
	})
}

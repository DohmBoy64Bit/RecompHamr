package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

const initTimeout = 30 * time.Second

type Client struct {
	name    string
	version string

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	mu       sync.Mutex
	reqID    int64
	pending  map[int64]chan *Response
	readerDone chan struct{}

	serverInfo ServerInfo
	tools      []ToolDef
}

func NewClient(name, version string) *Client {
	return &Client{
		name:    name,
		version: version,
	}
}

func (c *Client) Connected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cmd != nil && c.cmd.Process != nil
}

func (c *Client) Name() string    { return c.name }
func (c *Client) Version() string { return c.version }

func (c *Client) ServerName() string {
	if c.serverInfo.Name != "" {
		return c.serverInfo.Name
	}
	return c.name
}

func (c *Client) Tools() []ToolDef {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]ToolDef, len(c.tools))
	copy(out, c.tools)
	return out
}

func (c *Client) Connect(ctx context.Context, command string, args ...string) error {
	c.mu.Lock()
	if c.cmd != nil {
		c.mu.Unlock()
		return fmt.Errorf("mcp %s: already connected", c.name)
	}
	// Use context.Background() for the process lifespan; the caller's ctx
	// (e.g. a slash-command timeout) must not kill the long-running bridge.
	c.cmd = exec.CommandContext(context.Background(), command, args...)
	c.cmd.Stderr = nil
	c.pending = make(map[int64]chan *Response)
	c.readerDone = make(chan struct{})
	c.mu.Unlock()

	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		c.cmd = nil
		return fmt.Errorf("mcp %s: stdin pipe: %w", c.name, err)
	}
	c.stdout, err = c.cmd.StdoutPipe()
	if err != nil {
		c.cmd = nil
		return fmt.Errorf("mcp %s: stdout pipe: %w", c.name, err)
	}

	if err := c.cmd.Start(); err != nil {
		c.cmd = nil
		return fmt.Errorf("mcp %s: start: %w", c.name, err)
	}

	go c.readLoop()

	initCtx, cancel := context.WithTimeout(ctx, initTimeout)
	defer cancel()

	result, err := c.call(initCtx, "initialize", InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    ClientCapabilities{},
		ClientInfo: ClientInfo{
			Name:    c.name,
			Version: c.version,
		},
	})
	if err != nil {
		c.Disconnect()
		return fmt.Errorf("mcp %s: initialize: %w", c.name, err)
	}

	var initResult InitializeResult
	if err := json.Unmarshal(*result, &initResult); err != nil {
		c.Disconnect()
		return fmt.Errorf("mcp %s: parse initialize result: %w", c.name, err)
	}
	c.serverInfo = initResult.ServerInfo

	c.notify("notifications/initialized", nil)

	toolsResult, err := c.call(initCtx, "tools/list", nil)
	if err != nil {
		c.Disconnect()
		return fmt.Errorf("mcp %s: tools/list: %w", c.name, err)
	}

	var listResult ListToolsResult
	if err := json.Unmarshal(*toolsResult, &listResult); err != nil {
		c.Disconnect()
		return fmt.Errorf("mcp %s: parse tools/list result: %w", c.name, err)
	}

	c.mu.Lock()
	c.tools = listResult.Tools
	c.mu.Unlock()

	return nil
}

func (c *Client) Disconnect() {
	c.mu.Lock()
	if c.cmd == nil || c.cmd.Process == nil {
		c.mu.Unlock()
		return
	}
	c.stdin.Close()
	proc := c.cmd.Process
	c.cmd = nil
	c.mu.Unlock()

	if proc != nil {
		proc.Kill()
	}
}

func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (*CallToolResult, error) {
	raw, err := c.call(ctx, "tools/call", CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return nil, err
	}
	var result CallToolResult
	if err := json.Unmarshal(*raw, &result); err != nil {
		return nil, fmt.Errorf("mcp %s: parse tool call result: %w", c.name, err)
	}
	return &result, nil
}

func (c *Client) call(ctx context.Context, method string, params interface{}) (*json.RawMessage, error) {
	c.mu.Lock()
	if c.cmd == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("mcp %s: not connected", c.name)
	}
	id := atomic.AddInt64(&c.reqID, 1)
	ch := make(chan *Response, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	req := Request{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	c.mu.Lock()
	_, err = c.stdin.Write(data)
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("mcp %s: write: %w", c.name, err)
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("mcp %s: rpc error %d: %s", c.name, resp.Error.Code, resp.Error.Message)
		}
		if resp.Result == nil {
			return nil, fmt.Errorf("mcp %s: %q returned null result", c.name, method)
		}
		return resp.Result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) notify(method string, params interface{}) {
	notif := Notification{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  params,
	}
	data, _ := json.Marshal(notif)
	data = append(data, '\n')
	c.mu.Lock()
	c.stdin.Write(data)
	c.mu.Unlock()
}

func (c *Client) readLoop() {
	defer close(c.readerDone)
	scanner := bufio.NewScanner(c.stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var resp Response
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			continue
		}
		if resp.ID == 0 {
			continue
		}
		c.mu.Lock()
		ch, ok := c.pending[resp.ID]
		c.mu.Unlock()
		if ok {
			ch <- &resp
		}
	}
}

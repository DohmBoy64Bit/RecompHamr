package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const httpInitTimeout = 30 * time.Second

// HTTPClient is an MCP client that connects via streamable-http transport.
type HTTPClient struct {
	name   string
	version string

	baseURL string
	http    *http.Client

	mu         sync.Mutex
	reqID      int64
	serverInfo ServerInfo
	tools      []ToolDef
	connected  bool
}

func NewHTTPClient(name, version string) *HTTPClient {
	return &HTTPClient{
		name:    name,
		version: version,
		http: &http.Client{},
	}
}

func (c *HTTPClient) Name() string       { return c.name }
func (c *HTTPClient) Version() string    { return c.version }
func (c *HTTPClient) Connected() bool    { c.mu.Lock(); defer c.mu.Unlock(); return c.connected }
func (c *HTTPClient) ServerName() string {
	c.mu.Lock()
	if c.serverInfo.Name != "" {
		name := c.serverInfo.Name
		c.mu.Unlock()
		return name
	}
	c.mu.Unlock()
	return c.name
}

func (c *HTTPClient) Tools() []ToolDef {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]ToolDef, len(c.tools))
	copy(out, c.tools)
	return out
}

func (c *HTTPClient) Connect(ctx context.Context, baseURL string) error {
	c.mu.Lock()
	if c.connected {
		c.mu.Unlock()
		return fmt.Errorf("mcp %s: already connected", c.name)
	}
	c.baseURL = stringsTrimRight(baseURL, "/")
	c.mu.Unlock()

	initCtx, cancel := context.WithTimeout(ctx, httpInitTimeout)
	defer cancel()

	// Initialize
	initReq := Request{
		JSONRPC: JSONRPCVersion,
		ID:      atomic.AddInt64(&c.reqID, 1),
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    ClientCapabilities{},
			ClientInfo: ClientInfo{
				Name:    c.name,
				Version: c.version,
			},
		},
	}
	initResult, err := c.post(initCtx, initReq)
	if err != nil {
		return fmt.Errorf("mcp %s: initialize: %w", c.name, err)
	}
	var initRes InitializeResult
	if err := json.Unmarshal(*initResult, &initRes); err != nil {
		return fmt.Errorf("mcp %s: parse initialize: %w", c.name, err)
	}

	c.mu.Lock()
	c.serverInfo = initRes.ServerInfo
	c.mu.Unlock()

	// Notify initialized
	notif := Notification{
		JSONRPC: JSONRPCVersion,
		Method:  "notifications/initialized",
	}
	c.sendNotification(initCtx, notif)

	// List tools
	toolsReq := Request{
		JSONRPC: JSONRPCVersion,
		ID:      atomic.AddInt64(&c.reqID, 1),
		Method:  "tools/list",
	}
	toolsResult, err := c.post(initCtx, toolsReq)
	if err != nil {
		return fmt.Errorf("mcp %s: tools/list: %w", c.name, err)
	}
	var listRes ListToolsResult
	if err := json.Unmarshal(*toolsResult, &listRes); err != nil {
		return fmt.Errorf("mcp %s: parse tools/list: %w", c.name, err)
	}

	c.mu.Lock()
	c.tools = listRes.Tools
	c.connected = true
	c.mu.Unlock()

	return nil
}

func (c *HTTPClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
}

func (c *HTTPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (*CallToolResult, error) {
	req := Request{
		JSONRPC: JSONRPCVersion,
		ID:      atomic.AddInt64(&c.reqID, 1),
		Method:  "tools/call",
		Params: CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}
	raw, err := c.post(ctx, req)
	if err != nil {
		return nil, err
	}
	var result CallToolResult
	if err := json.Unmarshal(*raw, &result); err != nil {
		return nil, fmt.Errorf("mcp %s: parse tool call: %w", c.name, err)
	}
	return &result, nil
}

func (c *HTTPClient) post(ctx context.Context, req Request) (*json.RawMessage, error) {
	c.mu.Lock()
	base := c.baseURL
	c.mu.Unlock()

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/mcp", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("mcp %s: HTTP %d: %s", c.name, resp.StatusCode, string(msg))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResp Response
	if err := json.Unmarshal(data, &rpcResp); err != nil {
		return nil, fmt.Errorf("mcp %s: parse response: %w", c.name, err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("mcp %s: rpc error %d: %s", c.name, rpcResp.Error.Code, rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

func (c *HTTPClient) sendNotification(ctx context.Context, notif Notification) {
	c.mu.Lock()
	base := c.baseURL
	c.mu.Unlock()

	body, _ := json.Marshal(notif)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, base+"/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	if resp, err := c.http.Do(httpReq); err == nil {
		resp.Body.Close()
	}
}

func stringsTrimRight(s, cutset string) string {
	for len(s) > 0 && containsRune(cutset, rune(s[len(s)-1])) {
		s = s[:len(s)-1]
	}
	return s
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

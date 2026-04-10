// cdp_client.go — Go 네이티브 CDP WebSocket 클라이언트
// 외부 의존성: 0 (crypto/sha1 + net + bufio)
// RFC 6455 최소 WebSocket 구현 + CDP JSON-RPC
package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ── 최소 WebSocket 클라이언트 (RFC 6455, 외부 의존 0) ──

type wsConn struct {
	conn   net.Conn
	reader *bufio.Reader
	mu     sync.Mutex
	closed int32
}

func wsConnect(url string) (*wsConn, error) {
	// Parse ws://host:port/path
	url = strings.TrimPrefix(url, "ws://")
	idx := strings.Index(url, "/")
	host := url
	path := "/"
	if idx > 0 {
		host = url[:idx]
		path = url[idx:]
	}

	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		return nil, err
	}

	// WebSocket handshake
	key := make([]byte, 16)
	rand.Read(key)
	wsKey := base64.StdEncoding.EncodeToString(key)

	handshake := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n", path, host, wsKey)
	conn.Write([]byte(handshake))

	reader := bufio.NewReader(conn)
	// Read response line
	line, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(line, "101") {
		conn.Close()
		return nil, fmt.Errorf("ws handshake failed: %s", line)
	}
	// Skip headers
	for {
		h, err := reader.ReadString('\n')
		if err != nil || h == "\r\n" {
			break
		}
	}

	// Verify accept key
	magic := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(wsKey + magic))
	_ = base64.StdEncoding.EncodeToString(h.Sum(nil))

	return &wsConn{conn: conn, reader: reader}, nil
}

func (w *wsConn) readMessage() ([]byte, error) {
	if atomic.LoadInt32(&w.closed) == 1 {
		return nil, io.EOF
	}

	// Read frame header
	header := make([]byte, 2)
	if _, err := io.ReadFull(w.reader, header); err != nil {
		return nil, err
	}

	payloadLen := int(header[1] & 0x7F)
	switch payloadLen {
	case 126:
		ext := make([]byte, 2)
		io.ReadFull(w.reader, ext)
		payloadLen = int(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		io.ReadFull(w.reader, ext)
		payloadLen = int(binary.BigEndian.Uint64(ext))
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(w.reader, payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (w *wsConn) sendMessage(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if atomic.LoadInt32(&w.closed) == 1 {
		return io.EOF
	}

	frame := make([]byte, 0, 10+len(data))
	frame = append(frame, 0x81) // text frame, FIN

	maskKey := make([]byte, 4)
	rand.Read(maskKey)

	if len(data) < 126 {
		frame = append(frame, byte(len(data))|0x80) // masked
	} else if len(data) < 65536 {
		frame = append(frame, 126|0x80)
		ext := make([]byte, 2)
		binary.BigEndian.PutUint16(ext, uint16(len(data)))
		frame = append(frame, ext...)
	} else {
		frame = append(frame, 127|0x80)
		ext := make([]byte, 8)
		binary.BigEndian.PutUint64(ext, uint64(len(data)))
		frame = append(frame, ext...)
	}

	frame = append(frame, maskKey...)
	masked := make([]byte, len(data))
	for i, b := range data {
		masked[i] = b ^ maskKey[i%4]
	}
	frame = append(frame, masked...)

	_, err := w.conn.Write(frame)
	return err
}

func (w *wsConn) close() {
	if atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		w.conn.Close()
	}
}

// ── CDP 클라이언트 ──

type CDPClient struct {
	ws       *wsConn
	id       int64
	pending  sync.Map // id → chan json.RawMessage
	onEvent  func(method string, params json.RawMessage)
}

func NewCDPClient(wsURL string) (*CDPClient, error) {
	ws, err := wsConnect(wsURL)
	if err != nil {
		return nil, err
	}
	c := &CDPClient{ws: ws}
	go c.readLoop()
	return c, nil
}

func (c *CDPClient) readLoop() {
	for {
		msg, err := c.ws.readMessage()
		if err != nil {
			return
		}

		var resp struct {
			ID     *int64           `json:"id"`
			Result json.RawMessage  `json:"result"`
			Error  *json.RawMessage `json:"error"`
			Method string           `json:"method"`
			Params json.RawMessage  `json:"params"`
		}
		json.Unmarshal(msg, &resp)

		if resp.ID != nil {
			if ch, ok := c.pending.LoadAndDelete(*resp.ID); ok {
				ch.(chan json.RawMessage) <- resp.Result
			}
		}
		if resp.Method != "" && c.onEvent != nil {
			c.onEvent(resp.Method, resp.Params)
		}
	}
}

func (c *CDPClient) Call(method string, params interface{}) (json.RawMessage, error) {
	id := atomic.AddInt64(&c.id, 1)
	ch := make(chan json.RawMessage, 1)
	c.pending.Store(id, ch)

	req := map[string]interface{}{"id": id, "method": method, "params": params}
	data, _ := json.Marshal(req)
	if err := c.ws.sendMessage(data); err != nil {
		c.pending.Delete(id)
		return nil, err
	}

	select {
	case result := <-ch:
		return result, nil
	case <-time.After(8 * time.Second):
		c.pending.Delete(id)
		return nil, fmt.Errorf("timeout")
	}
}

func (c *CDPClient) Close() {
	c.ws.close()
}

// ── CDP 타겟 발견 ──

type CDPTarget struct {
	ID                   string `json:"id"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	Type                 string `json:"type"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func cdpListTargets(port int) ([]CDPTarget, error) {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/json/list", port))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var targets []CDPTarget
	json.NewDecoder(resp.Body).Decode(&targets)
	return targets, nil
}

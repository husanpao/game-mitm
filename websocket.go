package gamemitm

import (
	"crypto/tls"
	"encoding/hex"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Session struct {
	Client *websocket.Conn
	Server *websocket.Conn
}

func (s *Session) SendTextToServer(data []byte) {
	s.Server.WriteMessage(websocket.TextMessage, data)
}
func (s *Session) SendBinaryToServer(data []byte) {
	s.Server.WriteMessage(websocket.BinaryMessage, data)
}
func (s *Session) SendTextToClient(data []byte) {
	s.Server.WriteMessage(websocket.TextMessage, data)
}
func (s *Session) SendBinaryToClient(data []byte) {
	s.Server.WriteMessage(websocket.BinaryMessage, data)
}

// handleWebSocket handles WebSocket connections
func (p *ProxyServer) handleWebSocket(w http.ResponseWriter, r *http.Request, isSecure bool) {
	scheme := "ws"
	if isSecure {
		scheme = "wss"
	}

	// 构建目标URL
	targetURL := url.URL{
		Scheme:   scheme,
		Host:     r.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	// 创建一个新的header对象，只复制需要的，避免WebSocket特定的头
	requestHeader := http.Header{}

	// 复制非WebSocket特定的头
	for k, vs := range r.Header {
		// 跳过WebSocket特定的头，让dialer自己添加这些
		if k != "Sec-Websocket-Extensions" &&
			k != "Sec-Websocket-Key" &&
			k != "Sec-Websocket-Version" &&
			k != "Sec-Websocket-Protocol" &&
			k != "Sec-Websocket-Accept" &&
			k != "Upgrade" &&
			k != "Connection" {
			requestHeader[k] = vs
		}
	}

	// 只复制子协议头，这个是必要的
	if proto := r.Header.Get("Sec-Websocket-Protocol"); proto != "" {
		requestHeader.Set("Sec-Websocket-Protocol", proto)
	}

	// 连接目标WebSocket服务器
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 增加超时和更详细的错误处理
	targetConn, resp, err := dialer.Dial(targetURL.String(), requestHeader)
	if err != nil {
		p.logger.Error("Failed to connect to target WebSocket server: %v ", err)
		if resp != nil {

			// 转发响应
			for k, vs := range resp.Header {
				for _, v := range vs {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
		} else {
			http.Error(w, "Unable to connect to WebSocket server", http.StatusBadGateway)
		}
		return
	}
	defer targetConn.Close()

	// Define client connection upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		Subprotocols: websocket.Subprotocols(r),
	}

	// Upgrade connection with client
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		p.logger.Error("Failed to upgrade WebSocket connection: %v", err)
		return
	}
	defer clientConn.Close()

	ctx := &ProxyCtx{
		Req:  r,
		Resp: resp,
		WSSession: &Session{
			Client: clientConn,
			Server: targetConn,
		},
	}
	for u, handle := range p.connectedHandles {
		if u == All || strings.Contains(r.Host, u) {
			if handle != nil {
				handle([]byte{}, ctx)
			}
		}
	}
	// Create channels for relaying messages
	clientDone := make(chan struct{})
	targetDone := make(chan struct{})

	// Forward messages from client to target server
	go func() {
		defer close(clientDone)
		for {
			messageType, message, err := clientConn.ReadMessage()
			if err != nil {
				p.logger.Error("Failed to read client message: %v", err)
				return
			}

			if p.Verbose {
				p.logger.Debug("Client -> Server: %s", hex.EncodeToString(message))
			}
			// Here you can add code to modify WebSocket messages

			modifiedMessage := message
			for u, handle := range p.reqHandles {
				if u == All || strings.Contains(r.Host, u) {
					if handle != nil {
						modifiedMessage = handle(message, ctx)
					}
				}
			}

			if err := targetConn.WriteMessage(messageType, modifiedMessage); err != nil {
				p.logger.Error("Failed to send message to target server: %v", err)
				return
			}
		}
	}()

	// Forward messages from target server to client
	go func() {
		defer close(targetDone)
		for {
			messageType, message, err := targetConn.ReadMessage()
			if err != nil {
				p.logger.Error("Failed to read target server message: %v", err)
				return
			}
			if p.Verbose {
				p.logger.Debug("Server -> Client: %s", hex.EncodeToString(message))
			}
			// Here you can add code to modify WebSocket messages

			modifiedMessage := message

			for u, handle := range p.respHandles {

				if u == All || strings.Contains(r.Host, u) {
					if handle != nil {
						modifiedMessage = handle(message, ctx)
					}
				}
			}
			if err := clientConn.WriteMessage(messageType, modifiedMessage); err != nil {
				p.logger.Error("Failed to send message to client: %v", err)
				return
			}
		}
	}()

	// Wait for either connection to close
	select {
	case <-clientDone:
		p.logger.Info("Client connection closed")
	case <-targetDone:
		p.logger.Info("Target server connection closed")
	}
}

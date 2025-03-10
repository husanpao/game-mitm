package gamemitm

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"net/http"
	"strings"
)

// proxyHTTPS handles HTTPS request/response cycle
func (p *ProxyServer) proxyHTTPS(clientConn *tls.Conn, destConn *tls.Conn, host string) {
	// Read client request
	httpReader := bufio.NewReader(clientConn)
	req, err := http.ReadRequest(httpReader)
	if err != nil {
		p.logger.Error("Failed to read client request from %s: %v", host, err)
		return
	}
	ctx := &ProxyCtx{
		Req: req,
	}

	// 检测是否为WebSocket升级请求
	if websocket.IsWebSocketUpgrade(req) {
		if p.Verbose {
			p.logger.Debug("Handling WebSocket (WSS) connection for %s", host)
		}

		rwAdapter := newTLSResponseWriter(clientConn)
		p.handleWebSocket(rwAdapter, req, true)
		return
	}

	// Read request body
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		p.logger.Error("Failed to read request body for %s: %v", host, err)
		return
	}
	req.Body.Close()
	modifiedReqBody := reqBody
	for url, handle := range p.reqHandles {
		if url == All || strings.Contains(req.Host, url) {
			if handle != nil {
				modifiedReqBody = handle(reqBody, ctx)
			}
		}
	}

	// Create new request to target server
	outReq, err := http.NewRequest(req.Method, "https://"+host+req.URL.String(), bytes.NewReader(modifiedReqBody))
	if err != nil {
		p.logger.Error("Failed to create request for %s: %v", host, err)
		return
	}

	// Copy request headers
	for key, values := range req.Header {
		for _, value := range values {
			outReq.Header.Add(key, value)
		}
	}

	// Send request to target server
	outReq.Write(destConn)

	// Read response from target server
	respReader := bufio.NewReader(destConn)
	resp, err := http.ReadResponse(respReader, outReq)
	if err != nil {
		p.logger.Error("Failed to read server response for %s: %v", host, err)
		return
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Error("Failed to read response body for %s: %v", host, err)
		return
	}
	resp.Body.Close()
	ctx.Resp = resp
	modifiedRespBody := respBody
	for url, handle := range p.respHandles {
		if url == All || strings.Contains(req.Host, url) {
			if handle != nil {
				modifiedRespBody = handle(respBody, ctx)
			}
		}
	}

	// Create new response to send to client
	outResp := &http.Response{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		Proto:         resp.Proto,
		ProtoMajor:    resp.ProtoMajor,
		ProtoMinor:    resp.ProtoMinor,
		Header:        resp.Header,
		Body:          io.NopCloser(bytes.NewReader(modifiedRespBody)),
		ContentLength: int64(len(modifiedRespBody)),
	}

	// Send response to client
	outResp.Write(clientConn)
}

type tlsResponseWriter struct {
	conn       *tls.Conn
	header     http.Header
	statusCode int
}

func newTLSResponseWriter(conn *tls.Conn) *tlsResponseWriter {
	return &tlsResponseWriter{
		conn:       conn,
		header:     make(http.Header),
		statusCode: http.StatusOK,
	}
}

func (tw *tlsResponseWriter) Header() http.Header {
	return tw.header
}

func (tw *tlsResponseWriter) Write(data []byte) (int, error) {
	return tw.conn.Write(data)
}

func (tw *tlsResponseWriter) WriteHeader(statusCode int) {
	tw.statusCode = statusCode
}

func (tw *tlsResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return tw.conn, bufio.NewReadWriter(bufio.NewReader(tw.conn), bufio.NewWriter(tw.conn)), nil
}

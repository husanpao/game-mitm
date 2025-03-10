package gamemitm

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// handleHTTP handles HTTP requests
func (p *ProxyServer) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// 创建目标URL
	targetURL := *r.URL
	targetURL.Scheme = "http"
	targetURL.Host = r.Host
	if r.URL.Host != "" {
		targetURL.Host = r.URL.Host
	}

	ctx := &ProxyCtx{
		Req: r,
	}

	// 读取请求体
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		p.logger.Error("Failed to read request body for %s: %v", r.URL, err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	modifiedReqBody := reqBody
	for url, handle := range p.reqHandles {
		if url == All || strings.Contains(r.Host, url) {
			if handle != nil {
				modifiedReqBody = handle(reqBody, ctx)
			}
		}
	}

	// 创建新的请求发送到目标服务器
	req, err := http.NewRequest(r.Method, targetURL.String(), bytes.NewReader(modifiedReqBody))
	if err != nil {
		p.logger.Error("Failed to create request for %s: %v", targetURL.String(), err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// 复制原始请求头部
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求到目标服务器
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		p.logger.Error("Failed to send request to target server %s: %v", targetURL.String(), err)
		http.Error(w, "Failed to send request to target server", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Error("Failed to read response body for %s: %v", targetURL.String(), err)
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}
	ctx.Resp = resp
	modifiedRespBody := respBody
	for url, handle := range p.respHandles {
		if url == All || strings.Contains(r.Host, url) {
			if handle != nil {
				modifiedRespBody = handle(respBody, ctx)
			}
		}
	}

	// 复制响应头部到客户端
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 设置响应状态码
	w.WriteHeader(resp.StatusCode)

	// 写入修改后的响应体
	_, err = w.Write(modifiedRespBody)
	if err != nil {
		p.logger.Error("Failed to write modified response body for %s: %v", r.URL, err)
	}
}

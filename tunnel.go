package gamemitm

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// handleTunneling handles HTTPS tunnel requests
func (p *ProxyServer) handleTunneling(w http.ResponseWriter, r *http.Request) {
	// 修复主机名格式
	host := r.Host
	if host == "" {
		p.logger.Error("Invalid Host header in the request")
		http.Error(w, "Invalid Host header", http.StatusBadRequest)
		return
	}
	// 清理多余的斜杠
	for len(host) > 0 && host[0] == '/' {
		host = host[1:]
	}

	// Hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		p.logger.Error("Hijacking not supported for this connection")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	// Send 200 OK to client
	w.WriteHeader(http.StatusOK)

	// Get client connection
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		p.logger.Error("Failed to hijack connection: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	// Get certificate for this domain
	cert, err := p.certManager.GetCertificateForDomain(host)
	if err != nil {
		p.logger.Error("Failed to generate certificate for %s: %v", host, err)
		return
	}

	// Create TLS config for client connection
	config := &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}

	// Create TLS connection with client
	tlsConn := tls.Server(clientConn, config)
	if err := tlsConn.Handshake(); err != nil {
		p.logger.Error("TLS handshake with client failed for %s: %v", host, err)
		return
	}
	defer tlsConn.Close()

	// Create dialer with timeout
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	// Connect to destination server
	destConn, err := dialer.Dial("tcp", host)
	if err != nil {
		p.logger.Error("Failed to connect to target server %s: %v", host, err)
		return
	}
	defer destConn.Close()
	serverName := host
	if h, _, err := net.SplitHostPort(serverName); err == nil {
		serverName = h
	}

	// Establish TLS connection to target server
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         serverName,
	}

	destTLSConn := tls.Client(destConn, tlsConfig)
	if err := destTLSConn.Handshake(); err != nil {
		p.logger.Error("TLS handshake with target server %s failed: %v", host, err)
		return
	}
	defer destTLSConn.Close()

	// Process HTTPS requests
	p.proxyHTTPS(tlsConn, destTLSConn, host)
}

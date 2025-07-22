package gamemitm

import (
	"context"
	"fmt"
	"github.com/husanpao/game-mitm/cert"
	"net/http"
	"os"
	"time"
)

type ProxyServer struct {
	logger             Logger
	port               int
	ca                 *cert.CA
	certManager        *cert.CertificateManager
	Verbose            bool
	reqHandles         map[string]Handle
	hasReqHandle       bool
	respHandles        map[string]Handle
	hasRespHandle      bool
	connectedHandles   map[string]Handle
	hasConnectedHandle bool
	server             *http.Server
}

func NewProxy() *ProxyServer {
	if err := os.MkdirAll("./ca", 0755); err != nil {
		panic(err)
	}
	ca, err := cert.LoadOrCreateCA("./ca")
	if err != nil {
		panic(err)
	}
	return &ProxyServer{
		logger:           NewDefaultLogger(),
		port:             12311,
		ca:               ca,
		certManager:      cert.NewCertificateManager(ca),
		Verbose:          true,
		reqHandles:       make(map[string]Handle),
		respHandles:      make(map[string]Handle),
		connectedHandles: make(map[string]Handle),
	}
}

func (p *ProxyServer) SetLogger(logger Logger) {
	p.logger = logger
}
func (p *ProxyServer) SetPort(port int) {
	p.port = port
}
func (p *ProxyServer) SetVerbose(verbose bool) {
	p.Verbose = verbose
}
func (p *ProxyServer) SetCa(ca *cert.CA) {
	p.ca = ca
	p.certManager = cert.NewCertificateManager(ca)
}

func (p *ProxyServer) Start() error {
	p.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", p.port),
		Handler: http.HandlerFunc(p.handleRequest),
	}
	p.logger.Info("Starting proxy server on port %d ", p.port)
	return p.server.ListenAndServe()
}

// Stop gracefully stops the proxy server with a timeout
func (p *ProxyServer) Stop() error {
	if p.server != nil {
		p.logger.Info("Stopping proxy server on port %d", p.port)
		// 创建一个带有超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 优雅地关闭服务器，等待活跃连接完成
		err := p.server.Shutdown(ctx)
		if err != nil {
			p.logger.Error("Server shutdown error: %v", err)
			// 如果优雅关闭失败，强制关闭
			return p.server.Close()
		}
		p.logger.Info("Server stopped gracefully")
	}
	return nil
}

func (p *ProxyServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Handle the incoming request and forward it to the target server
	if p.Verbose {
		p.logger.Debug("Received request: %s %s", r.Method, r.URL)
	}

	if r.Method == http.MethodConnect {
		if p.Verbose {
			p.logger.Debug("Handling CONNECT request for %s", r.URL)
		}
		p.handleTunneling(w, r)
		return
	}
	// 处理普通 HTTP 请求
	if p.Verbose {
		p.logger.Debug("Handling HTTP request for %s", r.URL)
	}
	p.handleHTTP(w, r)
}

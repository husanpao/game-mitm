package gamemitm

import (
	"fmt"
	"gamemitm/cert"
	"net/http"
	"os"
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
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", p.port),
		Handler: http.HandlerFunc(p.handleRequest),
	}
	p.logger.Info("Starting proxy server on port %d ", p.port)
	return server.ListenAndServe()
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

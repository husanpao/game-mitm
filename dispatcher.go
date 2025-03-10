package gamemitm

const (
	All = "*"
)
const (
	Request = iota + 1000
	Response
	Connected
)

type Dispatcher struct {
	handleType int
	url        string
	p          *ProxyServer
}

func NewDispatcher(handleType int, url string, p *ProxyServer) *Dispatcher {
	return &Dispatcher{handleType: handleType, url: url, p: p}
}
func (d *Dispatcher) Do(f Handle) {
	switch d.handleType {
	case Request:
		d.p.reqHandles[d.url] = f
	case Response:
		d.p.respHandles[d.url] = f
	case Connected:
		d.p.connectedHandles[d.url] = f
	}
}
func (p *ProxyServer) OnRequest(url string) *Dispatcher {
	d := NewDispatcher(Request, url, p)
	if p.hasReqHandle {
		p.logger.Warn("request handle [*] already exists")
		return d
	}
	if url == All {
		p.hasReqHandle = true
		p.reqHandles = make(map[string]Handle)
	}
	p.reqHandles[url] = nil
	return d
}

func (p *ProxyServer) OnResponse(url string) *Dispatcher {
	d := NewDispatcher(Response, url, p)
	if p.hasRespHandle {
		p.logger.Warn("response handle [*] already exists.")
		return d
	}
	if url == All {
		p.hasRespHandle = true
		p.respHandles = make(map[string]Handle)
	}
	p.respHandles[url] = nil
	return d
}

func (p *ProxyServer) OnConnected(url string) *Dispatcher {
	d := NewDispatcher(Connected, url, p)
	if p.hasConnectedHandle {
		p.logger.Warn("connected handle [*] already exists")
		return d
	}
	if url == All {
		p.hasConnectedHandle = true
		p.connectedHandles = make(map[string]Handle)
	}
	p.connectedHandles[url] = nil
	return d
}

package gamemitm

import "net/http"

type ProxyCtx struct {
	Req       *http.Request
	Resp      *http.Response
	WSSession *Session
	UserData  any
	Proxy     *ProxyServer
}
type Handle func(body []byte, ctx *ProxyCtx) []byte

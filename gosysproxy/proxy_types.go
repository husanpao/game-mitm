// proxy_types.go
package gosysproxy

type ProxyStatus struct {
	Type                 uint32
	Proxy                string
	Bypass               []string
	DisableProxyIntranet bool
}

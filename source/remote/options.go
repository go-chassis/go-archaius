package remote

import "crypto/tls"

//Options is client option
type Options struct {
	ServerURI     string
	Endpoint      string
	TLSConfig     *tls.Config
	TenantName    string
	EnableSSL     bool
	APIVersion    string
	AutoDiscovery bool
	RefreshPort   string

	Labels map[string]string
}

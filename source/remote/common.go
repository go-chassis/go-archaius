package remote

import (
	"crypto/tls"
	"errors"
	"time"
)

// errors
var (
	ErrInvalidEP      = errors.New("invalid endpoint")
	ErrLabelsNil      = errors.New("labels can not be nil")
	ErrAppEmpty       = errors.New("app can not be empty")
	ErrServiceTooLong = errors.New("exceeded max value for service name")
)

// const
const (
	LabelService     = "service"
	LabelVersion     = "version"
	LabelEnvironment = "environment"
	LabelApp         = "app"

	DefaultInterval = time.Second * 30
)

// Mode
const (
	ModeWatch = iota
	ModeInterval
)

// Options is client option
type Options struct {
	ServerURI     string
	Endpoint      string
	TLSConfig     *tls.Config
	TenantName    string
	EnableSSL     bool
	APIVersion    string
	AutoDiscovery bool
	RefreshPort   string
	WatchTimeOut  int
	VerifyPeer    bool
	ProjectID     string

	Labels map[string]string
}

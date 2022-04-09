package http

import (
	"crypto/tls"
	"net/http"
	"time"

	kilncommon "github.com/skillz-blockchain/go-utils/common"
	kilntls "github.com/skillz-blockchain/go-utils/crypto/tls"
	kilnnet "github.com/skillz-blockchain/go-utils/net"

	"golang.org/x/net/http2"
)

// TransportConfig options to configure communication between Traefik and the servers
type TransportConfig struct {
	Dialer                *kilnnet.DialerConfig
	IdleConnTimeout       *kilncommon.Duration
	ResponseHeaderTimeout *kilncommon.Duration
	ExpectContinueTimeout *kilncommon.Duration
	MaxIdleConnsPerHost   int
	MaxConnsPerHost       int
	DisableKeepAlives     bool
	DisableCompression    bool
	EnableHTTP2           bool

	TLS *kilntls.Config
}

func (cfg *TransportConfig) SetDefault() *TransportConfig {
	if cfg.Dialer == nil {
		cfg.Dialer = new(kilnnet.DialerConfig)
	}
	cfg.Dialer.SetDefault()

	if cfg.IdleConnTimeout == nil {
		cfg.IdleConnTimeout = &kilncommon.Duration{Duration: 90 * time.Second}
	}

	if cfg.ResponseHeaderTimeout == nil {
		cfg.ResponseHeaderTimeout = &kilncommon.Duration{Duration: 0}
	}

	if cfg.ExpectContinueTimeout == nil {
		cfg.ExpectContinueTimeout = &kilncommon.Duration{Duration: time.Second}
	}

	return cfg
}

// NewTransport creates a http.Transport
func NewTransport(cfg *TransportConfig) (*http.Transport, error) {
	// Create dialer
	dlr := kilnnet.NewDialer(cfg.Dialer)

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dlr.DialContext,
		DisableKeepAlives:     cfg.DisableKeepAlives,
		DisableCompression:    cfg.DisableCompression,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:       cfg.MaxConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout.Duration,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout.Duration,
		ExpectContinueTimeout: cfg.ExpectContinueTimeout.Duration,
	}

	if cfg.TLS != nil {
		tlsCfg, err := cfg.TLS.ToTLSConfig()
		if err != nil {
			return nil, err
		}

		tlsDlr := tls.Dialer{
			NetDialer: dlr,
			Config:    tlsCfg,
		}

		transport.DialTLSContext = tlsDlr.DialContext
	}

	if cfg.EnableHTTP2 {
		err := http2.ConfigureTransport(transport)
		if err != nil {
			return nil, err
		}
	}

	return transport, nil
}

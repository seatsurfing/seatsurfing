package util

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/config"
)

func GetDNSResolver() *net.Resolver {
	if GetConfig().DNSServer != "" {
		return &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Millisecond * time.Duration(5000),
				}
				return d.DialContext(ctx, network, GetConfig().DNSServer)
			},
		}
	}
	return &net.Resolver{}
}

func GetHTTPClientWithCustomDNS(insecureSkipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{
				Resolver:  GetDNSResolver(),
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext(ctx, network, addr)
		},
	}
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}
	return client
}

// GetHTTPClientInsecureSkipVerify returns an HTTP client that skips TLS
// certificate verification but uses the system/cluster DNS resolver.
// This is intentionally separate from GetHTTPClientWithCustomDNS: the custom
// DNS resolver is appropriate for DNS lookups (e.g. TXT records) but not for
// HTTP requests that must be routed through the cluster network (e.g. domain
// accessibility checks via an ingress controller).
func GetHTTPClientInsecureSkipVerify() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}
}

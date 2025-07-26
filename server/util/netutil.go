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
				Resolver: GetDNSResolver(),
			}).DialContext(ctx, network, addr)
		},
	}
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}
	return client
}

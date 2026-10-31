package util

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"syscall"
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
// The target host is derived from a customer-controlled domain, so with
// ALLOW_PRIVATE_OUTBOUND unset the resolved address is checked at connect time,
// which also covers redirects and DNS rebinding. Certificate verification is
// skipped because the check only evaluates the returned JSON payload and
// may run before a certificate has been issued for the domain.
func GetHTTPClientInsecureSkipVerify() *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	if !GetConfig().AllowPrivateOutbound {
		dialer.Control = blockedOutboundControl
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext:     dialer.DialContext,
	}
	return &http.Client{
		Timeout:       time.Second * 10,
		Transport:     tr,
		CheckRedirect: limitRedirects,
	}
}

func limitRedirects(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return errors.New("stopped after 5 redirects")
	}
	return nil
}

var ErrBlockedOutboundAddress = errors.New("destination address is not allowed")

var blockedOutboundNets = mustParseCIDRs(
	"0.0.0.0/8",
	"100.64.0.0/10",
	"192.0.0.0/24",
	"198.18.0.0/15",
	"240.0.0.0/4",
	"64:ff9b::/96",
	"64:ff9b:1::/48",
)

func mustParseCIDRs(cidrs ...string) []*net.IPNet {
	res := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(err)
		}
		res = append(res, n)
	}
	return res
}

func IsBlockedOutboundIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return true
	}
	for _, n := range blockedOutboundNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func blockedOutboundControl(network, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if ip == nil || IsBlockedOutboundIP(ip) {
		return ErrBlockedOutboundAddress
	}
	return nil
}

// GetOutboundHTTPClient returns a client for requests to user-configured URLs.
// Unless ALLOW_PRIVATE_OUTBOUND is set, the resolved address is checked at connect time,
// which also covers redirects and DNS rebinding.
func GetOutboundHTTPClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	if !GetConfig().AllowPrivateOutbound {
		dialer.Control = blockedOutboundControl
		tr.Proxy = nil
	}
	tr.DialContext = dialer.DialContext
	return &http.Client{
		Timeout:       15 * time.Second,
		Transport:     tr,
		CheckRedirect: limitRedirects,
	}
}

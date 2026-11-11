package linkcheck

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"
)

// Result is one address probe. Alive includes sites that answer but refuse
// bots. HardDead is a confirmed miss: unknown host, connection refused, 404
// or 410. Anything else is left alone so a timeout does not unpublish a site.
type Result struct {
	Alive     bool
	HardDead  bool
	Final     string
	Permanent bool
	Status    int
	Reason    string
}

var errPrivate = errors.New("地址指向内网")

// Check fetches a public http(s) URL. Private, loopback and link-local
// targets are refused and are not treated as a dead catalog entry.
func Check(raw string) Result {
	target := trimURL(raw)
	parsed, err := url.Parse(target)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return Result{Reason: "不是网页地址"}
	}
	if parsed.User != nil {
		return Result{Reason: "地址里带了账号"}
	}
	if ip := net.ParseIP(parsed.Hostname()); ip != nil && !publicIP(ip) {
		return Result{Reason: errPrivate.Error()}
	}

	var permanent bool
	dialer := &net.Dialer{Timeout: 8 * time.Second}
	transport := &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialPublic(ctx, dialer, network, addr)
		},
		TLSHandshakeTimeout:   8 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		DisableKeepAlives:     true,
		MaxIdleConns:          0,
	}
	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("跳转太多")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return errors.New("跳到了非网页地址")
			}
			if req.URL.User != nil {
				return errors.New("跳转地址带了账号")
			}
			prev := via[len(via)-1].Response
			if prev != nil && (prev.StatusCode == http.StatusMovedPermanently || prev.StatusCode == http.StatusPermanentRedirect) {
				permanent = true
			}
			return nil
		},
	}
	req, err := http.NewRequest(http.MethodGet, parsed.String(), nil)
	if err != nil {
		return Result{Reason: "地址无法请求"}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KaiwuluLinkCheck/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		alive, hard, reason := classify(0, err)
		return Result{Alive: alive, HardDead: hard, Reason: reason}
	}
	defer resp.Body.Close()
	_, _ = io.CopyN(io.Discard, resp.Body, 1024)

	final := ""
	if resp.Request != nil && resp.Request.URL != nil {
		u := *resp.Request.URL
		u.Fragment = ""
		final = u.String()
	}
	alive, hard, reason := classify(resp.StatusCode, nil)
	return Result{
		Alive:     alive,
		HardDead:  hard,
		Final:     final,
		Permanent: permanent,
		Status:    resp.StatusCode,
		Reason:    reason,
	}
}

func dialPublic(ctx context.Context, dialer *net.Dialer, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, &net.DNSError{Err: "no such host", Name: host, IsNotFound: true}
	}
	for _, ip := range ips {
		if !publicIP(ip.IP) {
			return nil, errPrivate
		}
	}
	return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
}

func publicIP(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 0 || (v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127) {
			return false
		}
	}
	return true
}

func classify(status int, err error) (alive, hard bool, reason string) {
	if err != nil {
		if errors.Is(err, errPrivate) {
			return false, false, errPrivate.Error()
		}
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
			return false, true, "域名不存在"
		}
		if errors.Is(err, syscall.ECONNREFUSED) {
			return false, true, "连接被拒绝"
		}
		return false, false, "暂时连不上"
	}
	switch {
	case status >= 200 && status < 300:
		return true, false, ""
	case status == http.StatusUnauthorized || status == http.StatusForbidden || status == http.StatusMethodNotAllowed || status == http.StatusTooManyRequests:
		return true, false, ""
	case status == http.StatusNotFound || status == http.StatusGone:
		return false, true, "页面不存在"
	default:
		return false, false, "暂时打不开"
	}
}

func httpsUpgrade(from, to string) bool {
	a, err1 := url.Parse(from)
	b, err2 := url.Parse(to)
	if err1 != nil || err2 != nil {
		return false
	}
	if a.Scheme != "http" || b.Scheme != "https" {
		return false
	}
	if !strings.EqualFold(a.Hostname(), b.Hostname()) {
		return false
	}
	return a.EscapedPath() == b.EscapedPath() && a.RawQuery == b.RawQuery
}

func trimURL(raw string) string {
	return strings.TrimSpace(raw)
}

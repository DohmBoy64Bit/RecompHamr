package tools

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// RecompReferenceName is the stable model-facing reference-cache tool name.
const RecompReferenceName = "recomp_reference"

const maxReferenceBytes = 512 << 10

// RecompReferenceSchema returns the stable OpenAI-compatible function definition.
func RecompReferenceSchema() map[string]any {
	return map[string]any{"type": "function", "function": map[string]any{
		"name":        RecompReferenceName,
		"description": "Fetch a public HTTP(S) reference page into the project's private 24-hour cache. HTML is reduced to readable text; use read_file on the returned path. Private and local network destinations are refused.",
		"parameters": map[string]any{"type": "object", "properties": map[string]any{
			"url": map[string]any{"type": "string", "description": "Public HTTP or HTTPS reference URL without embedded credentials."},
		}, "required": []string{"url"}},
	}}
}

func validateReferenceURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil {
		return nil, errors.New("expected a credential-free public http(s) URL")
	}
	if strings.EqualFold(u.Hostname(), "localhost") || net.ParseIP(strings.Trim(u.Hostname(), "[]")) != nil && !publicIP(net.ParseIP(strings.Trim(u.Hostname(), "[]"))) {
		return nil, errors.New("local or non-public destination refused")
	}
	return u, nil
}

func publicIP(ip net.IP) bool {
	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	if !address.IsGlobalUnicast() {
		return false
	}
	for _, prefix := range nonPublicPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

var nonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"), netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"), netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"), netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"), netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("100::/64"), netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("fc00::/7"), netip.MustParsePrefix("fe80::/10"),
}

func newPublicHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	return newPublicHTTPClientWith(net.DefaultResolver.LookupIP, dialer.DialContext)
}

func newPublicHTTPClientWith(resolve func(context.Context, string, string) ([]net.IP, error), dial func(context.Context, string, string) (net.Conn, error)) *http.Client {
	transport := &http.Transport{DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := resolve(ctx, "ip", host)
		if err != nil {
			return nil, err
		}
		for _, ip := range ips {
			if !publicIP(ip) {
				return nil, errors.New("recomp_reference: DNS resolved to a non-public address")
			}
		}
		if len(ips) == 0 {
			return nil, errors.New("recomp_reference: DNS returned no addresses")
		}
		return dial(ctx, network, net.JoinHostPort(ips[0].String(), port))
	}}
	client := &http.Client{Transport: transport, Timeout: 15 * time.Second}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many redirects")
		}
		_, err := validateReferenceURL(req.URL.String())
		return err
	}
	return client
}

func (s Set) recompReference(ctx context.Context, rawURL string) string {
	if s.privateRoot == "" || s.restrict == nil || s.httpClient == nil || s.now == nil {
		return "(recomp_reference: output directory not configured)"
	}
	u, err := validateReferenceURL(rawURL)
	if err != nil {
		return "(recomp_reference: invalid URL: " + err.Error() + ")"
	}
	cacheRoot := filepath.Join(s.privateRoot, "reference")
	if err := preparePrivateDir(cacheRoot, s.restrict); err != nil {
		return "(recomp_reference: cache: " + err.Error() + ")"
	}
	name := strings.Trim(sanitizeCacheName(u.Hostname()+"-"+strings.Trim(u.EscapedPath(), "/")), "-_.")
	if len(name) > 96 {
		name = name[:96]
	}
	digest := sha256.Sum256([]byte(u.String()))
	name += fmt.Sprintf("-%x", digest[:6])
	outPath := filepath.Join(cacheRoot, name+".txt")
	if info, err := cacheLstat(outPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return "(recomp_reference: unsafe cache entry refused)"
		}
		if age := s.now().Sub(info.ModTime()); age >= 0 && age < 24*time.Hour {
			if err := s.restrict(outPath, false); err != nil {
				return "(recomp_reference: secure cache: " + err.Error() + ")"
			}
			return fmt.Sprintf("Cached (from %s)\n  → %s\n\nUse read_file to inspect.", info.ModTime().Format("2006-01-02 15:04"), outPath)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "(recomp_reference: cache: " + err.Error() + ")"
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	req.Header.Set("Accept", "text/html, text/plain;q=0.9, application/json;q=0.8, */*;q=0.1")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "(recomp_reference: cancelled)"
		}
		return "(recomp_reference: fetch failed: " + boundedDiagnostic([]byte(err.Error())) + ")"
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("(recomp_reference: HTTP %d from %s)", resp.StatusCode, u.Hostname())
	}
	body, err := readBoundedBody(resp.Body, maxReferenceBytes)
	if err != nil {
		return "(recomp_reference: read: " + err.Error() + ")"
	}
	text := string(body)
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/html") {
		text = extractHTMLText(body)
	}
	now := s.now()
	displayURL := redactedReferenceURL(u)
	content := fmt.Sprintf("# Source: %s\n# Fetched: %s\n\n%s", displayURL, now.Format("2006-01-02 15:04"), text)
	if err := atomicPrivateWrite(outPath, []byte(content), s.restrict); err != nil {
		return "(recomp_reference: write: " + err.Error() + ")"
	}
	return fmt.Sprintf("Fetched %s\n  → %s · %s\n\nUse read_file to inspect. Cached for 24 hours.", displayURL, outPath, humanSize(int64(len(content))))
}

func readBoundedBody(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response exceeds %s", humanSize(limit))
	}
	return data, nil
}

func extractHTMLText(data []byte) string {
	doc, _ := html.Parse(strings.NewReader(string(data)))
	var out strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "script", "style", "nav", "footer", "header", "noscript":
				return
			}
			switch node.Data {
			case "br", "p", "div", "li", "tr", "h1", "h2", "h3", "h4", "h5", "h6":
				out.WriteByte('\n')
			}
		}
		if node.Type == html.TextNode {
			if text := strings.TrimSpace(node.Data); text != "" {
				out.WriteString(text)
				out.WriteByte(' ')
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return strings.TrimSpace(out.String())
}

func sanitizeCacheName(value string) string {
	var out strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			out.WriteRune(r)
		} else {
			out.WriteByte('_')
		}
	}
	return out.String()
}

func redactedReferenceURL(value *url.URL) string {
	copy := *value
	if copy.RawQuery != "" {
		copy.RawQuery = "redacted"
	}
	return copy.String()
}

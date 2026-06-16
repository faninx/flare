package fn

import (
	"net"
	"net/http"
	"regexp"
	"strings"
)

// DynamicURL holds parsed request URL components. Use ParseRequestURLTo and ParseDynamicUrlWith to avoid global state.
type DynamicURL struct {
	Host     string
	Hostname string
	Href     string
	Origin   string
	Pathname string
	Port     string
	Protocol string
}

// RequestURL is the package-level parsed URL (set by ParseRequestURL). Prefer ParseRequestURLTo and passing *DynamicURL for concurrency-safe use.
var RequestURL DynamicURL

var hostPortRe = regexp.MustCompile(`([\w+\.-]+):(\d+)$`)

func getPort(host string, defaultPort string) (hostname string, port string) {
	hostname = host
	port = defaultPort
	portMatch := hostPortRe.FindStringSubmatch(host)
	if portMatch != nil {
		hostname = portMatch[1]
		port = portMatch[2]
	}
	return
}

// ParseRequestURLTo parses r into a DynamicURL without using package-level state. Prefer this over ParseRequestURL when possible.
func ParseRequestURLTo(r *http.Request) DynamicURL {
	scheme := "http:"
	defaultPort := "80"
	if r != nil && r.TLS != nil {
		scheme = "https:"
		defaultPort = "443"
	}
	host := ""
	if r != nil {
		host = r.Host
	}
	hostname, port := getPort(host, defaultPort)
	pathname := ""
	requestURI := ""
	if r != nil && r.URL != nil {
		pathname = r.URL.Path
		requestURI = r.URL.RequestURI()
	}
	return DynamicURL{
		Host:     host,
		Hostname: hostname,
		Href:     strings.Join([]string{scheme, "//", host, requestURI}, ""),
		Origin:   strings.Join([]string{scheme, "//", host}, ""),
		Pathname: pathname,
		Port:     port,
		Protocol: scheme,
	}
}

// ParseRequestURL parses r and updates package-level RequestURL. For new code, prefer ParseRequestURLTo and passing *DynamicURL.
func ParseRequestURL(r *http.Request) {
	RequestURL = ParseRequestURLTo(r)
}

// ParseDynamicUrlWith substitutes URL placeholders using info. Concurrency-safe when info is request-scoped.
func ParseDynamicUrlWith(url string, info *DynamicURL) string {
	if info == nil {
		return url
	}
	result := url
	result = strings.ReplaceAll(result, "{host}", info.Host)
	result = strings.ReplaceAll(result, "{hostname}", info.Hostname)
	result = strings.ReplaceAll(result, "{href}", info.Href)
	result = strings.ReplaceAll(result, "{origin}", info.Origin)
	result = strings.ReplaceAll(result, "{pathname}", info.Pathname)
	result = strings.ReplaceAll(result, "{port}", info.Port)
	result = strings.ReplaceAll(result, "{protocol}", info.Protocol)
	return result
}

func ParseDynamicUrl(url string) string {
	return ParseDynamicUrlWith(url, &RequestURL)
}

// EnvMode describes the user's preferred network environment for bookmark URL resolution.
type EnvMode int

const (
	EnvAuto EnvMode = iota
	EnvLAN
	EnvWAN
)

// String returns a stable lowercase identifier ("auto", "lan", "wan") for cookie and template use.
func (e EnvMode) String() string {
	switch e {
	case EnvLAN:
		return "lan"
	case EnvWAN:
		return "wan"
	default:
		return "auto"
	}
}

// ParseEnvMode normalizes a cookie/querystring value into an EnvMode. Unknown values fall back to EnvAuto.
func ParseEnvMode(s string) EnvMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "lan", "local", "private", "intranet":
		return EnvLAN
	case "wan", "public", "external", "internet":
		return EnvWAN
	default:
		return EnvAuto
	}
}

// IsLANHost reports whether the given host (with optional :port) belongs to a private network.
// Rules:
//   - IPv4 in 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 127.0.0.0/8
//   - IPv6 ::1, fc00::/7 (unique local addresses)
//   - Hostname suffixes: .lan, .local, .internal, .home, .intranet
//   - "localhost"
// Returns false for empty input or unrecognised values.
func IsLANHost(host string) bool {
	if host == "" {
		return false
	}
	hostname, _ := getPort(host, "")
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return false
	}
	if strings.EqualFold(hostname, "localhost") {
		return true
	}
	if ip := net.ParseIP(hostname); ip != nil {
		return isPrivateIP(ip)
	}
	lower := strings.ToLower(hostname)
	if strings.HasSuffix(lower, ".lan") ||
		strings.HasSuffix(lower, ".local") ||
		strings.HasSuffix(lower, ".internal") ||
		strings.HasSuffix(lower, ".home") ||
		strings.HasSuffix(lower, ".intranet") {
		return true
	}
	return false
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}
	if ip.IsPrivate() {
		return true
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if ip.IsUnspecified() {
		return true
	}
	return false
}

// ResolveBookmarkURL returns the final URL for a bookmark under the current environment and request.
// Behavior matrix:
//   - env=lan  : always use link
//   - env=wan  : use linkPublic; falls back to link if linkPublic is empty
//   - env=auto : use link when the request host is a LAN host, otherwise use linkPublic (falling back to link if empty)
//
// The chosen value is then run through ParseDynamicUrlWith so {host} / {hostname} / etc. placeholders still work.
func ResolveBookmarkURL(link, linkPublic string, env EnvMode, info *DynamicURL) string {
	var target string
	switch env {
	case EnvLAN:
		target = link
	case EnvWAN:
		if linkPublic != "" {
			target = linkPublic
		} else {
			target = link
		}
	default:
		if linkPublic == "" || (info != nil && IsLANHost(info.Host)) {
			target = link
		} else {
			target = linkPublic
		}
	}
	return ParseDynamicUrlWith(target, info)
}

// ShouldHideInWanMode reports whether a bookmark/app should be omitted from the rendered page
// because the user has explicitly chosen WAN mode but the item has no public URL configured.
//
// Why this is a separate function rather than baking the behavior into ResolveBookmarkURL:
//   - URL resolution (which link to follow) and visibility (whether to render the item at all)
//     are different concerns. Keeping them split means the existing 6+ ResolveBookmarkURL
//     unit tests stay valid and the render layer can decide per-context (e.g. show in
//     search, hide in grid) without changing the resolver.
//   - Hiding an item whose only link would resolve to a 192.168.x.x address is a UX
//     safeguard: in WAN mode the resolver falls back to `link` when `linkPublic` is empty,
//     and clicking that from a public-network client is guaranteed to fail.
//
// Only EnvWAN is treated as "public-network context" by this helper: readEnvMode in
// internal/pages/home/env.go collapses EnvAuto (and unknown cookie values) back to EnvLAN,
// so in practice this is the only branch that ever reaches the renderers.
func ShouldHideInWanMode(linkPublic string, env EnvMode) bool {
	return env == EnvWAN && linkPublic == ""
}

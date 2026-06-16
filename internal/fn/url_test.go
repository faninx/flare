package fn

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestParseRequestURL_HTTP(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com:8080/foo/bar", nil)
	r.Host = "example.com:8080"
	ParseRequestURL(r)
	if RequestURL.Host != "example.com:8080" {
		t.Errorf("Host: got %q", RequestURL.Host)
	}
	if RequestURL.Hostname != "example.com" {
		t.Errorf("Hostname: got %q", RequestURL.Hostname)
	}
	if RequestURL.Port != "8080" {
		t.Errorf("Port: got %q", RequestURL.Port)
	}
	if RequestURL.Protocol != "http:" {
		t.Errorf("Protocol: got %q", RequestURL.Protocol)
	}
	if RequestURL.Pathname != "/foo/bar" {
		t.Errorf("Pathname: got %q", RequestURL.Pathname)
	}
}

func TestParseRequestURL_HTTPS(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	r.Host = "example.com"
	r.TLS = &tls.ConnectionState{}
	ParseRequestURL(r)
	if RequestURL.Protocol != "https:" {
		t.Errorf("Protocol: got %q", RequestURL.Protocol)
	}
	if RequestURL.Port != "443" {
		t.Errorf("Port: got %q", RequestURL.Port)
	}
}

func TestParseRequestURL_NoPort(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	r.Host = "example.com"
	ParseRequestURL(r)
	if RequestURL.Hostname != "example.com" || RequestURL.Port != "80" {
		t.Errorf("Hostname=%q Port=%q", RequestURL.Hostname, RequestURL.Port)
	}
}

func TestParseDynamicUrl(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "http://localhost:5005/", nil)
	r.Host = "localhost:5005"
	ParseRequestURL(r)
	out := ParseDynamicUrl("origin={origin} host={host} path={pathname}")
	if out != "origin=http://localhost:5005 host=localhost:5005 path=/" {
		t.Errorf("ParseDynamicUrl: got %q", out)
	}
	out2 := ParseDynamicUrl("no placeholders")
	if out2 != "no placeholders" {
		t.Errorf("ParseDynamicUrl no placeholders: got %q", out2)
	}
}

func TestIsLANHost(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		// RFC 1918 IPv4
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.254", true},
		{"192.168.0.1", true},
		{"192.168.1.100", true},
		{"192.168.50.10", true},

		// loopback
		{"127.0.0.1", true},
		{"127.1.2.3", true},

		// link-local
		{"169.254.1.1", true},

		// public IPv4 — not LAN
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"172.32.0.1", false},
		{"172.15.255.255", false},
		{"11.0.0.1", false},
		{"193.168.1.1", false},

		// IPv6
		{"::1", true},
		{"fd00::1", true},
		{"fdff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", true},
		{"fe80::1", true},
		{"2001:db8::1", false},

		// hostnames
		{"localhost", true},
		{"LOCALHOST", true},
		{"home.lan", true},
		{"nas.home.lan", true},
		{"router.local", true},
		{"server.internal", true},
		{"office.intranet", true},
		{"example.com", false},
		{"cloudflare.com", false},
		{"flare.ddns.example.com", false},
		{"flare.lan.example.com", false}, // ".lan" suffix is true
		{"notalan", false},               // not a suffix
		{"lan.com", false},               // not a suffix

		// host:port
		{"192.168.1.10:5005", true},
		{"example.com:443", false},

		// empty
		{"", false},
	}
	for _, tc := range cases {
		got := IsLANHost(tc.host)
		if got != tc.want {
			t.Errorf("IsLANHost(%q) = %v, want %v", tc.host, got, tc.want)
		}
	}
}

func TestParseEnvMode(t *testing.T) {
	cases := []struct {
		in   string
		want EnvMode
	}{
		{"", EnvAuto},
		{"auto", EnvAuto},
		{"AUTO", EnvAuto},
		{"lan", EnvLAN},
		{"LAN", EnvLAN},
		{" local ", EnvLAN},
		{"private", EnvLAN},
		{"intranet", EnvLAN},
		{"wan", EnvWAN},
		{"WAN", EnvWAN},
		{"public", EnvWAN},
		{"internet", EnvWAN},
		{"garbage", EnvAuto},
	}
	for _, tc := range cases {
		got := ParseEnvMode(tc.in)
		if got != tc.want {
			t.Errorf("ParseEnvMode(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestEnvModeString(t *testing.T) {
	if EnvAuto.String() != "auto" {
		t.Errorf("EnvAuto.String() = %q", EnvAuto.String())
	}
	if EnvLAN.String() != "lan" {
		t.Errorf("EnvLAN.String() = %q", EnvLAN.String())
	}
	if EnvWAN.String() != "wan" {
		t.Errorf("EnvWAN.String() = %q", EnvWAN.String())
	}
}

func newReq(t *testing.T, host string) *http.Request {
	t.Helper()
	r, _ := http.NewRequest(http.MethodGet, "http://"+host+"/", nil)
	r.Host = host
	return r
}

func TestResolveBookmarkURL_AutoLAN(t *testing.T) {
	r := newReq(t, "192.168.1.10:5005")
	info := ParseRequestURLTo(r)
	got := ResolveBookmarkURL("http://nas.local:8096", "https://jellyfin.ddns.example.com", EnvAuto, &info)
	if got != "http://nas.local:8096" {
		t.Errorf("auto+LAN host should pick link; got %q", got)
	}
}

func TestResolveBookmarkURL_AutoWAN(t *testing.T) {
	r := newReq(t, "flare.ddns.example.com:5005")
	info := ParseRequestURLTo(r)
	got := ResolveBookmarkURL("http://192.168.1.20:8096", "https://jellyfin.ddns.example.com", EnvAuto, &info)
	if got != "https://jellyfin.ddns.example.com" {
		t.Errorf("auto+WAN host with link_public should pick link_public; got %q", got)
	}
}

func TestResolveBookmarkURL_AutoWANFallback(t *testing.T) {
	r := newReq(t, "flare.ddns.example.com:5005")
	info := ParseRequestURLTo(r)
	got := ResolveBookmarkURL("http://example.com/app", "", EnvAuto, &info)
	if got != "http://example.com/app" {
		t.Errorf("auto+WAN+empty link_public should fall back to link; got %q", got)
	}
}

func TestResolveBookmarkURL_ForceLAN(t *testing.T) {
	r := newReq(t, "flare.ddns.example.com:5005")
	info := ParseRequestURLTo(r)
	got := ResolveBookmarkURL("http://192.168.1.20:8096", "https://jellyfin.ddns.example.com", EnvLAN, &info)
	if got != "http://192.168.1.20:8096" {
		t.Errorf("force LAN should always pick link; got %q", got)
	}
}

func TestResolveBookmarkURL_ForceWAN(t *testing.T) {
	r := newReq(t, "192.168.1.10:5005")
	info := ParseRequestURLTo(r)
	got := ResolveBookmarkURL("http://192.168.1.20:8096", "https://jellyfin.ddns.example.com", EnvWAN, &info)
	if got != "https://jellyfin.ddns.example.com" {
		t.Errorf("force WAN should pick link_public; got %q", got)
	}
}

func TestResolveBookmarkURL_ForceWANFallback(t *testing.T) {
	r := newReq(t, "192.168.1.10:5005")
	info := ParseRequestURLTo(r)
	got := ResolveBookmarkURL("http://only-this-one.example.com", "", EnvWAN, &info)
	if got != "http://only-this-one.example.com" {
		t.Errorf("force WAN with empty link_public should fall back to link; got %q", got)
	}
}

func TestResolveBookmarkURL_PlaceholderStillWorks(t *testing.T) {
	r := newReq(t, "192.168.1.10:5005")
	info := ParseRequestURLTo(r)
	got := ResolveBookmarkURL("http://{hostname}:8096", "", EnvAuto, &info)
	if got != "http://192.168.1.10:8096" {
		t.Errorf("placeholders should still expand; got %q", got)
	}
}

// TestResolveBookmarkURL_NilInfo covers the safety net: callers that don't have a *DynamicURL
// (e.g. some test harnesses) must not panic. In Auto mode with link_public set, the function
// conservatively picks link_public since it cannot prove the request came from a LAN host.
func TestResolveBookmarkURL_NilInfo(t *testing.T) {
	if got := ResolveBookmarkURL("http://a", "https://b", EnvAuto, nil); got != "https://b" {
		t.Errorf("auto+nil info+link_public set should pick link_public; got %q", got)
	}
	if got := ResolveBookmarkURL("http://only", "", EnvAuto, nil); got != "http://only" {
		t.Errorf("auto+nil info+empty link_public should fall back to link; got %q", got)
	}
	if got := ResolveBookmarkURL("http://a", "https://b", EnvLAN, nil); got != "http://a" {
		t.Errorf("force LAN+nil info should pick link; got %q", got)
	}
	if got := ResolveBookmarkURL("http://a", "https://b", EnvWAN, nil); got != "https://b" {
		t.Errorf("force WAN+nil info should pick link_public; got %q", got)
	}
}

func TestShouldHideInWanMode(t *testing.T) {
	// The four env × linkPublic combinations. Hide is only true when the user has
	// explicitly chosen WAN mode and there is no public URL configured — the resolver
	// would otherwise silently fall back to a LAN-only link that fails from a public client.
	cases := []struct {
		name       string
		linkPublic string
		env        EnvMode
		want       bool
	}{
		{"wan+empty hides (the bug we are preventing)", "", EnvWAN, true},
		{"wan+public keeps visible", "https://example.com", EnvWAN, false},
		{"lan+empty keeps visible (LAN link works locally)", "", EnvLAN, false},
		{"lan+public keeps visible", "https://example.com", EnvLAN, false},
		{"auto+empty keeps visible (readEnvMode collapses auto to LAN, but the helper stays safe)", "", EnvAuto, false},
		{"auto+public keeps visible", "https://example.com", EnvAuto, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ShouldHideInWanMode(tc.linkPublic, tc.env); got != tc.want {
				t.Errorf("ShouldHideInWanMode(%q, %v) = %v, want %v", tc.linkPublic, tc.env, got, tc.want)
			}
		})
	}
}

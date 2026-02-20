package goodbots

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestNormalizeHost exercises trimming of trailing dots and joining.
func TestNormalizeHost(t *testing.T) {
	cases := []struct {
		in   []string
		sep  string
		want string
	}{
		{[]string{"a.", "b.", "c."}, ";", "a;b;c"},
		{[]string{"foo", "bar."}, ",", "foo,bar"},
		{[]string{}, "|", ""},
	}
	for _, c := range cases {
		got := normalizeHost(c.in, c.sep)
		if got != c.want {
			t.Errorf("normalizeHost(%#v,%q) = %q; want %q", c.in, c.sep, got, c.want)
		}
	}
}

// TestBotDomain checks our regex against known good and bad domains.
func TestBotDomain(t *testing.T) {
	cases := []struct {
		domain string
		want   bool
	}{
		{"google", true},
		{"googlebot", true},
		{"msn", true},
		{"yahoo", true},
		{"archive", true},
		{"naver", true},
		{"example", false},
		{"googlexyz", false},
		{"bot", false},
	}
	for _, c := range cases {
		if got := botDomain(c.domain); got != c.want {
			t.Errorf("botDomain(%q) = %v; want %v", c.domain, got, c.want)
		}
	}
}

// TestRoundRobinDNSServer verifies that len(dnsServers) consecutive calls to
// dnsServer() return all distinct servers (round-robin property).
func TestRoundRobinDNSServer(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < len(dnsServers); i++ {
		s := dnsServer()
		seen[s] = true
	}
	if len(seen) != len(dnsServers) {
		t.Errorf("expected %d distinct DNS servers across %d consecutive calls, got %d: %v",
			len(dnsServers), len(dnsServers), len(seen), seen)
	}
}

// TestReverseDNSLocalhost verifies that 127.0.0.1 reverses to something containing "localhost".
func TestReverseDNSLocalhost(t *testing.T) {
	hosts, err := ReverseDNS(context.Background(), "127.0.0.1")
	if err != nil {
		t.Skipf("skipping reverse-lookup test; got error: %v", err)
	}
	found := false
	for _, h := range hosts {
		if strings.HasPrefix(strings.TrimRight(h, "."), "localhost") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ReverseDNS(127.0.0.1) = %v; expected at least one entry starting with localhost", hosts)
	}
}

// TestForwardDNSLocalhost verifies that "localhost" forwards back to loopback.
func TestForwardDNSLocalhost(t *testing.T) {
	ip, err := ForwardDNS(context.Background(), "localhost")
	if err != nil {
		t.Skipf("skipping forward-lookup test; got error: %v", err)
	}
	if ip != "127.0.0.1" && ip != "::1" {
		t.Errorf("ForwardDNS(localhost) = %q; want 127.0.0.1 or ::1", ip)
	}
}

// TestContextTimeoutReverseDNS verifies that a cancelled context causes an error.
func TestContextTimeoutReverseDNS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately before the call
	_, err := ReverseDNS(ctx, "8.8.8.8")
	if err == nil {
		t.Error("expected error with cancelled context, got nil")
	}
}

// TestContextTimeoutForwardDNS verifies that a cancelled context causes an error.
func TestContextTimeoutForwardDNS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately before the call
	_, err := ForwardDNS(ctx, "google.com")
	if err == nil {
		t.Error("expected error with cancelled context, got nil")
	}
}

// TestLastLineNoNewline verifies that input without a trailing newline is processed.
func TestLastLineNoNewline(t *testing.T) {
	input := "127.0.0.1" // no trailing newline
	buf := &bytes.Buffer{}
	if err := ResolveNames(1, context.Background(), strings.NewReader(input), buf); err != nil {
		t.Fatalf("ResolveNames: %v", err)
	}
	got := buf.String()
	if !strings.HasPrefix(got, "127.0.0.1\t") {
		t.Errorf("ResolveNames output = %q; expected it to start with '127.0.0.1\\t'", got)
	}
}

// TestResolveNamesLocalhost is a minimal integration: 127.0.0.1 -> localhost (possibly among others).
func TestResolveNamesLocalhost(t *testing.T) {
	input := "127.0.0.1\n"
	buf := &bytes.Buffer{}
	if err := ResolveNames(1, context.Background(), strings.NewReader(input), buf); err != nil {
		t.Fatalf("ResolveNames: %v", err)
	}
	got := buf.String()
	// The output is "127.0.0.1\t<hosts>" where <hosts> may be semicolon-joined
	// (e.g. "localhost;othername") if the loopback has multiple PTR records.
	if !strings.HasPrefix(got, "127.0.0.1\t") {
		t.Errorf("ResolveNames output = %q; expected prefix '127.0.0.1\\t'", got)
	}
	hostField := strings.TrimRight(strings.TrimPrefix(got, "127.0.0.1\t"), "\n")
	found := false
	for _, h := range strings.Split(hostField, ";") {
		if strings.HasPrefix(h, "localhost") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ResolveNames output = %q; expected 'localhost' in host field", got)
	}
}

// TestConcurrentWritesNoInterleave runs ResolveNames with multiple IPs concurrently
// and verifies each output line is well-formed (has at least two tab-separated fields).
func TestConcurrentWritesNoInterleave(t *testing.T) {
	const count = 10
	var sb strings.Builder
	for i := 0; i < count; i++ {
		sb.WriteString("127.0.0.1\n")
	}
	buf := &bytes.Buffer{}
	if err := ResolveNames(5, context.Background(), strings.NewReader(sb.String()), buf); err != nil {
		t.Fatalf("ResolveNames: %v", err)
	}
	output := strings.TrimRight(buf.String(), "\n")
	lines := strings.Split(output, "\n")
	if len(lines) != count {
		t.Errorf("expected %d output lines, got %d", count, len(lines))
	}
	for _, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			t.Errorf("malformed output line (too few tab-separated fields): %q", line)
		}
	}
}

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

// TestReverseDNSLocalhost verifies that 127.0.0.1 reverses to something containing "localhost".
func TestReverseDNSLocalhost(t *testing.T) {
   hosts, err := ReverseDNS("127.0.0.1")
   if err != nil {
       t.Skipf("skipping reverse‐lookup test; got error: %v", err)
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
   ip, err := ForwardDNS("localhost")
   if err != nil {
       t.Skipf("skipping forward‐lookup test; got error: %v", err)
   }
   if ip != "127.0.0.1" && ip != "::1" {
       t.Errorf("ForwardDNS(localhost) = %q; want 127.0.0.1 or ::1", ip)
   }
}

// TestResolveNamesLocalhost is a minimal integration: 127.0.0.1 -> localhost
func TestResolveNamesLocalhost(t *testing.T) {
   input := "127.0.0.1\n"
   buf := &bytes.Buffer{}
   if err := ResolveNames(1, context.Background(), strings.NewReader(input), buf); err != nil {
       t.Fatalf("ResolveNames: %v", err)
   }
   got := buf.String()
   want := "127.0.0.1\tlocalhost\n"
   if got != want {
       t.Errorf("ResolveNames output = %q; want %q", got, want)
   }
}

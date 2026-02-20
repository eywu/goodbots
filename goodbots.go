package goodbots

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tld "github.com/weppos/publicsuffix-go/publicsuffix"
	"golang.org/x/sync/semaphore"
)

var (
	protocol = "udp"
	port     = "53"
)

var dnsServers = []string{"1.0.0.1", "1.1.1.1", "8.8.4.4", "8.8.8.8", "208.67.222.222", "208.67.220.220"}

var dnsCounter atomic.Uint64

// dnsServer returns DNS servers in round-robin order using an atomic counter.
func dnsServer() string {
	idx := dnsCounter.Add(1) - 1
	return dnsServers[idx%uint64(len(dnsServers))]
}

// resolver is the package-level DNS resolver, reused across all lookups.
// net.Resolver is concurrency-safe.
var resolver = &net.Resolver{
	PreferGo: true,
	Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, protocol, dnsServer()+":"+port)
	},
}

func normalizeHost(hosts []string, delimiter string) string {
	var host []string
	for _, h := range hosts {
		host = append(host, strings.TrimRight(h, "."))
	}
	return strings.Join(host, delimiter)
}

var botRe = regexp.MustCompile(`^(google|googlebot|msn|pinterest|yandex|baidu|coccoc|yahoo|archive|naver)$`)

func botDomain(domain string) bool {
	return botRe.MatchString(domain)
}

// ReverseDNS performs a reverse DNS lookup on ip with a 5-second timeout.
func ReverseDNS(ctx context.Context, ip string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return resolver.LookupAddr(ctx, ip)
}

// ForwardDNS performs a forward DNS lookup on host with a 5-second timeout.
func ForwardDNS(ctx context.Context, host string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	addrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil || len(addrs) == 0 {
		return "", fmt.Errorf("forward lookup %q: %w", host, err)
	}
	return addrs[0].IP.String(), nil
}

// ResolveNames reads IPs from r, resolves each via reverse DNS, and writes
// results to w. A single writer goroutine serialises all output through a
// bufio.Writer, preventing concurrent writes and buffering output efficiently.
func ResolveNames(cc int64, ctx context.Context, r io.Reader, w io.Writer) error {
	br := bufio.NewReader(r)
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(cc)

	results := make(chan string, cc)

	bw := bufio.NewWriter(w)
	var writerDone sync.WaitGroup
	writerDone.Add(1)
	go func() {
		defer writerDone.Done()
		for msg := range results {
			fmt.Fprint(bw, msg)
		}
		bw.Flush()
	}()

	processIP := func(ip string) {
		defer wg.Done()
		defer sem.Release(1)

		host, err := ReverseDNS(ctx, ip)
		if err != nil {
			results <- fmt.Sprintf("%s\t%s\t%s\n", ip, "(error)", err)
			return
		}
		if len(host) == 0 {
			results <- fmt.Sprintf("%s\t%s\n", ip, "(none)")
			return
		}
		results <- fmt.Sprintf("%s\t%s\n", ip, normalizeHost(host, ";"))
	}

	var readErr error
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				// Handle last line that has no trailing newline.
				if trimmed := strings.TrimSpace(line); trimmed != "" {
					if err2 := sem.Acquire(ctx, 1); err2 == nil {
						wg.Add(1)
						go processIP(trimmed)
					}
				}
			} else {
				readErr = err
			}
			break
		}
		if err2 := sem.Acquire(ctx, 1); err2 != nil {
			readErr = err2
			break
		}
		wg.Add(1)
		go processIP(strings.TrimSpace(line))
	}

	wg.Wait()
	close(results)
	writerDone.Wait()
	return readErr
}

// GoodBots reads IPs from r, verifies each is a known search-engine bot via
// reverse+forward DNS, and writes confirmed bots to w.
func GoodBots(cc int64, ctx context.Context, r io.Reader, w io.Writer) error {
	br := bufio.NewReader(r)
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(cc)

	results := make(chan string, cc)

	bw := bufio.NewWriter(w)
	var writerDone sync.WaitGroup
	writerDone.Add(1)
	go func() {
		defer writerDone.Done()
		for msg := range results {
			fmt.Fprint(bw, msg)
		}
		bw.Flush()
	}()

	processIP := func(ip string) {
		defer wg.Done()
		defer sem.Release(1)

		host, err := ReverseDNS(ctx, ip)
		if err != nil {
			return
		}
		if len(host) == 0 {
			return
		}
		h := strings.TrimRight(host[0], ".")
		hname, _ := tld.Parse(h)
		if !botDomain(hname.SLD) {
			return
		}
		ip2, _ := ForwardDNS(ctx, h)
		if ip == ip2 {
			results <- fmt.Sprintf("%s\t%s\n", ip, h)
		}
	}

	var readErr error
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				// Handle last line that has no trailing newline.
				if trimmed := strings.TrimSpace(line); trimmed != "" {
					if err2 := sem.Acquire(ctx, 1); err2 == nil {
						wg.Add(1)
						go processIP(trimmed)
					}
				}
			} else {
				readErr = err
			}
			break
		}
		if err2 := sem.Acquire(ctx, 1); err2 != nil {
			readErr = err2
			break
		}
		wg.Add(1)
		go processIP(strings.TrimSpace(line))
	}

	wg.Wait()
	close(results)
	writerDone.Wait()
	return readErr
}

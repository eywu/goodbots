package goodbots

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	tld "github.com/weppos/publicsuffix-go/publicsuffix"
	"golang.org/x/sync/semaphore"
)

var (
	protocol = "udp"
	port     = "53"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func dnsServer() string {
	dnsServer := []string{"1.0.0.1", "1.1.1.1", "8.8.4.4", "8.8.8.8", "208.67.222.222", "208.67.220.220"}
	return dnsServer[rand.Intn(len(dnsServer))]
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

func newResolver() *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, protocol, dnsServer()+":"+port)
		},
	}
}

func ReverseDNS(ip string) ([]string, error) {
	r := newResolver()
	return r.LookupAddr(context.Background(), ip)
}

func ForwardDNS(host string) (string, error) {
	r := newResolver()
	addrs, err := r.LookupIPAddr(context.Background(), host)
	if err != nil || len(addrs) == 0 {
		return "", fmt.Errorf("forward lookup %q: %w", host, err)
	}
	return addrs[0].IP.String(), nil
}

func ResolveNames(cc int64, ctx context.Context, r io.Reader, w io.Writer) error {
	br := bufio.NewReader(r)
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(cc)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		err = sem.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		wg.Add(1)

		go func(ip string) {
			defer wg.Done()
			defer sem.Release(1)

			host, err := ReverseDNS(ip)

			if err != nil {
				fmt.Fprintf(w, "%s\t%s\t%s\n", ip, "(error)", err)
				return
			}
			if len(host) == 0 {
				fmt.Fprintf(w, "%s\t%s\n", ip, "(none)")
				return
			}

			fmt.Fprintf(w, "%s\t%s\n", ip, normalizeHost(host, ";"))
		}(line)
	}
	wg.Wait()
	return nil
}

func GoodBots(cc int64, ctx context.Context, r io.Reader, w io.Writer) error {
	br := bufio.NewReader(r)
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(cc)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		err = sem.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		wg.Add(1)

		go func(ip string) {
			defer wg.Done()
			defer sem.Release(1)
			host, err := ReverseDNS(ip)

			if err != nil {
				return
			}
			if len(host) == 0 {
				return
			}
			h := strings.TrimRight(host[0], ".")
			hname, _ := tld.Parse(h)

			tf, _ := botDomain(hname.SLD)

			if !tf {
				return
			}

			ip2, err := ForwardDNS(h)

			if ip == ip2 {
				fmt.Fprintf(w, "%s\t%s\n", ip, h)
			}
		}(line)
	}
	wg.Wait()
	return nil
}

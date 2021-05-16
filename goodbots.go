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

func dnsServer() string {
	dnsServer := []string{"1.0.0.1", "1.1.1.1", "8.8.4.4", "8.8.8.8", "208.67.222.222", "208.67.220.220"}
	rand.Seed(time.Now().Unix())
	return dnsServer[rand.Intn(len(dnsServer))]
}

func normalizeHost(hosts []string, delimiter string) string {
	var host []string
	for _, h := range hosts {
		host = append(host, strings.TrimRight(h, "."))
	}
	return strings.Join(host, delimiter)
}

func botDomain(domain string) (bool, error) {
	y, err := regexp.Match(`^(google|googlebot|msn|pinterest|yandex|baidu|coccoc|yahoo|archive|naver)$`, []byte(domain))
	return y, err
}

func ReverseDNS(ip string) ([]string, error) {
	// https://stackoverflow.com/questions/59889882/specifying-dns-server-for-lookup-in-go
	var r *net.Resolver

	r = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, protocol, dnsServer()+":"+port)
		},
	}
	host, err := r.LookupAddr(context.Background(), ip)

	return host, err
}

func ForwardDNS(host string) (string, error) {
	var r *net.Resolver

	r = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, protocol, dnsServer()+":"+port)
		},
	}
	ip, err := r.LookupIPAddr(context.Background(), host)
	return ip[0].IP.String(), err
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

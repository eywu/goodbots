# goodbots - trust but verify

![goodbots ip lookup example on googlebot ips](https://gist.github.com/eywu/0a7c55cb70d84e8ed3bf070acc52b2ec/raw/b3174dbdaa543c7f55039fe9c69815bbbeb436b5/goodbots.gif)
goodbots verifies the IP addresses of respectful crawlers like Googlebot by performing [reverse dns](https://searchsignals.com/how-to-do-a-reverse-dns-lookup) and forward dns
lookups.

1. Given an IP address (ex. `66.249.87.225`)
2. It performs a reverse dns lookup to get a hostname (ex. `crawl-66-249-87-225.googlebot.com`)
3. Then does a forward dns lookup on the hostname to get an IP (ex. `66.249.87.225`)
4. It compares the 1st IP to the 2nd IP
5. If they match, goodbots outputs the IP and hostname

## The Job-to-be-Done (#jtbd)

In search engine optimization (SEO), it is common to analyze a site's access logs (aka bot logs). Often there are
various requests by spoofed user-agents pretending to be official search engine crawlers like [Googlebot](https://developers.google.com/search/docs/advanced/crawling/googlebot). In order to have an accurate understanding of the site's crawl rate, we want to verify the IP address of the various crawlers.

# Getting Started

## How to install/build goodbots

Clone the repo:

```
git clone git@github.com:eywu/goodbots.git
```

Change to the `/cmd/goodbots` directory:

```
cd goodbots/cmd/goodbots
```

Build the binary/executable `main.go` file:

```
go build
```

## How to use goodbots

If you've built the `main.go` file that comes with goodbots above, you can simply feed goodbots IPs via `standard-in`.

Test a single IP

```
echo "203.208.60.1" | ./goodbots
```

Test a range of IPs with [prips command line tool](http://manpages.ubuntu.com/manpages/bionic/man1/prips.1.html)

```
prips 203.208.40.1 203.208.80.1 | ./goodbots
```

Test a list of IPs from a text or csv file

```
./goodbots < ip-list.txt
```

**note:** The CSV or text file expects only an IP on its own line.

Example:

```
66.249.87.224
203.208.23.146
203.208.23.126
203.208.60.227
```

### Saving the results

goodbots prints to `standard-out` with tab (\t) delimiters, so you can capture the output with an [output redirect](https://www.codecademy.com/learn/learn-the-command-line/modules/learn-the-command-line-redirection/cheatsheet).

**Example Output**

```
203.208.60.1 crawl-203-208-60-1.googlebot.com
66.249.85.123 google-proxy-66-249-85-123.google.com
66.249.87.12 rate-limited-proxy-66-249-87-12.google.com
66.249.85.224 google-proxy-66-249-85-224.google.com
```

Save verified bot IPs provide in a file name `ip-list.txt` to a filed named `saved-results.tsv`

```
./goodbots < ip-list.txt > saved-results.tsv
```

## DNS Resolvers

goodbots randomly selects a different public DNS resolver for each DNS lookup to reduce the chances of being blocked or
throttled by your DNS provider if you have lots of IPs to verify.

It uses these DNS providers:

- [CloudFlare Public DNS](https://www.cloudflare.com/learning/dns/what-is-1.1.1.1/)
  - 1.1.1.1
  - 1.0.0.1
- [Google Public DNS](https://developers.google.com/speed/public-dns)
  - 8.8.8.8
  - 8.8.4.4
- [Open DNS](https://www.opendns.com/setupguide/)
  - 208.67.222.222
  - 208.67.220.220
- [Quad9 DNS](https://www.quad9.net/) (⛔ _not supported yet_)
  - ~~9.9.9.9~~
  - ~~149.112.112.112~~

## Supported Crawlers

Currently verifying the domain name is a little imprecise. goodbots looks for just the domain name to match and does
**NOT** match the TLD.

Future improvements will test for more precise domains based on the crawlers specifications.

- googlebot
  - .googlebot.
  - .google.
- msnbot
  - .msn.
- bingbot
  - .msn.
- pinterest
  - .pinterest.
- yandex
  - .yandex.
- baidu
  - .baidu.
- coccoc
  - .coccoc.

## Make it go faster!

By default we only set the concurrency of requests to 10. If you want to speed up the work, you can increase that
number by modifying the `main.go` file before building the binary/executable.

# Other usage of goodbots

In building goodbots, we created a general purpose function for simply resolving the hostnames of any IP address.

In `main.go` you can uncomment the line that calls `ResolveNames()` and comment out the `GoodBots()` function call.

This will not perform a forward DNS lookup to verify the hostname resolves to the same IP address. Additionally, it
will output errors to the TSV output when it encounters IPs that error out when requesting the hostname.

```
➜  goodbots git:(main) ✗ prips -i 50 66.100.0.0 66.200.0.0 | ./goodbots
66.100.0.50	(error)	lookup 50.0.100.66.in-addr.arpa. on 192.168.1.1:53: no such host
...
66.100.1.144	(error)	lookup 144.1.100.66.in-addr.arpa. on 192.168.1.1:53: no such host
66.100.0.150	WebGods
66.100.0.250	(error)	lookup 250.0.100.66.in-addr.arpa. on 192.168.1.1:53: no such host
...
66.100.4.76	(error)	lookup 76.4.100.66.in-addr.arpa. on 192.168.1.1:53: no such host
66.100.4.126	mail.esai.com
```

---

# Other Resources

- [Google Documentation on Verifying Googlebot](https://developers.google.com/search/docs/advanced/crawling/verifying-googlebot?hl=en)
  - [JSON feed of Googlebot IPv4 and IPv6 ranges](https://developers.google.com/search/apis/ipranges/googlebot.json)
- [Google published IP ranges for Google API + services](https://support.google.com/a/answer/10026322?hl=en) h/t [Michael Stapelberg](https://github.com/stapelberg)
  - [JSON feed](https://www.gstatic.com/ipranges/goog.json)
- [DuckDuckGo published IPs](https://help.duckduckgo.com/duckduckgo-help-pages/results/duckduckbot/)
  - [DuckAssistBot published IPs](https://duckduckgo.com/duckduckgo-help-pages/results/duckassistbot)
- [Facebook published IP ranges](https://developers.facebook.com/docs/sharing/webmasters/crawler/)
- [Pinterest: Verify pinterestbot](https://help.pinterest.com/en/business/article/pinterestbot)
- [Apple: Verify applebot](https://support.apple.com/en-us/119829)
- [Internet Archive: Verify archive.org_bot](http://crawler.archive.org/index.html)
- [Bidu: Verify baiduspider](https://help.baidu.com/question?prod_id=99&class=0&id=3001)
- [Yandex: Verify yadex crawlers](https://yandex.com/support/webmaster/en/robot-workings/check-yandex-robots.html)
- [Cốc Cốc: Verify coccocbot](https://coccoc.com/search/console/en/coc-coc-robots)
- [Yahoo: Verify slurp](https://help.yahoo.com/kb/search-for-desktop/SLN22600.html?impressions=true)

---

Written in [Golang](https://golang.org/)
![gopher](https://user-images.githubusercontent.com/185250/118390633-f73be280-b5e4-11eb-8f60-bba0abb2f119.png)
Gopher courtesy of [Gopherize.me](https://gopherize.me/)

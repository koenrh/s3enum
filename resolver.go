package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
)

type Stats struct {
	Checked  uint64
	Found    uint64
	Errors   uint64
	Timeouts uint64
	Refused  uint64
	Duration time.Duration
}

func (s Stats) Summary() string {
	summary := fmt.Sprintf("completed: %d checked, %d found", s.Checked, s.Found)

	if s.Errors > 0 {
		summary += fmt.Sprintf(", %d errors", s.Errors)

		var breakdown []string
		if s.Timeouts > 0 {
			breakdown = append(breakdown, fmt.Sprintf("%d timeout", s.Timeouts))
		}
		if s.Refused > 0 {
			breakdown = append(breakdown, fmt.Sprintf("%d refused", s.Refused))
		}
		if other := s.Errors - s.Timeouts - s.Refused; other > 0 {
			breakdown = append(breakdown, fmt.Sprintf("%d other", other))
		}
		summary += fmt.Sprintf(" (%s)", strings.Join(breakdown, ", "))
	}

	summary += fmt.Sprintf(" in %.1fs", s.Duration.Seconds())
	return summary
}

type Resolver interface {
	IsBucket(string) bool
	Stats() Stats
}

type DNSResolver struct {
	client   dns.Client
	config   dns.ClientConfig
	checked  uint64
	found    uint64
	errors   uint64
	timeouts uint64
	refused  uint64
}

// NewS3Resolver initializes a new S3Resolver
func NewDNSResolver(nsAddr string) (*DNSResolver, error) {
	config, err := getConfig(nsAddr)

	if err != nil {
		return nil, err
	}

	return &DNSResolver{
		client: dns.Client{ReadTimeout: 2 * time.Second},
		config: *config,
	}, nil
}

const (
	defaultPort    int    = 53
	s3GlobalSuffix string = "s3.amazonaws.com."
	s31WSuffix     string = "s3-1-w.amazonaws.com."
)

func (s *DNSResolver) Stats() Stats {
	return Stats{
		Checked:  atomic.LoadUint64(&s.checked),
		Found:    atomic.LoadUint64(&s.found),
		Errors:   atomic.LoadUint64(&s.errors),
		Timeouts: atomic.LoadUint64(&s.timeouts),
		Refused:  atomic.LoadUint64(&s.refused),
	}
}

// IsBucket determines whether this prefix is a valid S3 bucket name.
func (s *DNSResolver) IsBucket(name string) bool {
	atomic.AddUint64(&s.checked, 1)

	records, err := s.resolveName(fmt.Sprintf("%s.%s", name, s3GlobalSuffix))

	if err != nil {
		atomic.AddUint64(&s.errors, 1)
		s.classifyError(err)
		return false
	}

	if len(records) != 1 || records[0].Header().Rrtype != dns.TypeCNAME {
		return false
	}

	cname := records[0].(*dns.CNAME)

	// The assumption is that existing bucket names (under the 's3.amazonaws.com' suffix)
	// resolve to either a regional S3 name (e.g. 's3-w.eu-central-1.amazonaws.com.') or
	// 's3-2-w.amazonaws.com.' and 's3-3-w.amazonaws.com.'. Hence, we assume that any bucket
	// that does not resolve 's3-1-w.amazonaws.com.' is an existing bucket.
	found := cname.Target != s31WSuffix
	if found {
		atomic.AddUint64(&s.found, 1)
	}
	return found
}

func (s *DNSResolver) classifyError(err error) {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		atomic.AddUint64(&s.timeouts, 1)
		return
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if strings.Contains(opErr.Err.Error(), "connection refused") {
			atomic.AddUint64(&s.refused, 1)
			return
		}
	}
}

func getConfig(nameserver string) (*dns.ClientConfig, error) {
	if nameserver != "" {
		h, p := parseHostAndPort(nameserver)

		return &dns.ClientConfig{
			Servers: []string{h},
			Port:    strconv.Itoa(p),
		}, nil
	}

	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")

	if err != nil {
		return nil, errors.New("could not read local resolver config")
	}

	return config, nil
}

func parseHostAndPort(addr string) (host string, port int) {
	if p := strings.Split(addr, ":"); len(p) == 2 {
		host = p[0]
		var err error
		if port, err = strconv.Atoi(p[1]); err != nil {
			port = defaultPort
		}
	} else {
		host = addr
	}

	if port == 0 {
		port = defaultPort
	}
	return
}

func (s *DNSResolver) resolveName(name string) ([]dns.RR, error) {
	msg := dns.Msg{}
	msg.SetQuestion(name, dns.TypeCNAME)
	addr := net.JoinHostPort(s.config.Servers[0], s.config.Port)
	retries := 3
	delay := 1 * time.Second

	var err error
	var r *dns.Msg
	for attempt := 1; attempt <= retries; attempt++ {
		r, _, err = s.client.Exchange(&msg, addr)
		if err == nil {
			break
		}
		if attempt != retries {
			time.Sleep(1 * delay)
			delay *= 2
		}
	}

	if err != nil {
		return nil, err
	}

	var answer = r.Answer
	if len(answer) == 0 {
		return []dns.RR{}, errors.New("empty answer")
	}

	return answer, nil
}

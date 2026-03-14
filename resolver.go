package main

import (
	"context"
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
	IsBucket(ctx context.Context, name string) bool
	Stats() Stats
}

type DNSResolver struct {
	client   dns.Client
	addr     string
	connPool chan *dns.Conn
	checked  uint64
	found    uint64
	errors   uint64
	timeouts uint64
	refused  uint64
}

func NewDNSResolver(nsAddr string) (*DNSResolver, error) {
	config, err := getConfig(nsAddr)
	if err != nil {
		return nil, err
	}

	addr := net.JoinHostPort(config.Servers[0], config.Port)

	return &DNSResolver{
		client:   dns.Client{ReadTimeout: 2 * time.Second},
		addr:     addr,
		connPool: make(chan *dns.Conn, 128),
	}, nil
}

func (s *DNSResolver) getConn() (*dns.Conn, error) {
	select {
	case conn := <-s.connPool:
		return conn, nil
	default:
		return s.client.Dial(s.addr)
	}
}

func (s *DNSResolver) putConn(conn *dns.Conn) {
	select {
	case s.connPool <- conn:
	default:
		conn.Close()
	}
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
func (s *DNSResolver) IsBucket(ctx context.Context, name string) bool {
	atomic.AddUint64(&s.checked, 1)

	records, err := s.resolveName(ctx, fmt.Sprintf("%s.%s", name, s3GlobalSuffix))

	if err != nil {
		if ctx.Err() != nil {
			return false
		}
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
		host, port := parseHostPort(nameserver)
		return &dns.ClientConfig{
			Servers: []string{host},
			Port:    port,
		}, nil
	}

	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return nil, errors.New("could not read local resolver config")
	}

	return config, nil
}

func parseHostPort(addr string) (host, port string) {
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		return addr, strconv.Itoa(defaultPort)
	}
	return h, p
}

func (s *DNSResolver) resolveName(ctx context.Context, name string) ([]dns.RR, error) {
	msg := dns.Msg{}
	msg.SetQuestion(name, dns.TypeCNAME)
	retries := 3
	delay := 1 * time.Second

	var err error
	var r *dns.Msg
	for attempt := 1; attempt <= retries; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		conn, dialErr := s.getConn()
		if dialErr != nil {
			err = dialErr
		} else {
			r, _, err = s.client.ExchangeWithConn(&msg, conn)
			if err == nil {
				s.putConn(conn)
				break
			}
			conn.Close()
		}

		if attempt != retries {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			delay *= 2
		}
	}

	if err != nil {
		return nil, err
	}

	if len(r.Answer) == 0 {
		return nil, errors.New("empty answer")
	}

	return r.Answer, nil
}

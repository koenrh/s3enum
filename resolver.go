package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type Resolver interface {
	IsBucket(string) bool
}

type DNSResolver struct {
	client dns.Client
	config dns.ClientConfig
}

// NewS3Resolver initializes a new S3Resolver
func NewDNSResolver(nsAddr string) (*DNSResolver, error) {
	config, err := getConfig(nsAddr)

	if err != nil {
		return nil, err
	}

	return &DNSResolver{
		client: dns.Client{ReadTimeout: 3 * time.Second},
		config: *config,
	}, nil
}

const (
	defaultPort    int    = 53
	s3GlobalSuffix string = "s3.amazonaws.com."
	s31WSuffix     string = "s3-1-w.amazonaws.com."
)

// IsBucket determines whether this prefix is a valid S3 bucket name.
func (s *DNSResolver) IsBucket(name string) bool {
	records, err := s.resolveName(fmt.Sprintf("%s.%s", name, s3GlobalSuffix))

	if err != nil {
		// TODO: Handle error
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
	return cname.Target != s31WSuffix
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

	var answer = r.Answer
	if len(answer) == 0 {
		return []dns.RR{}, errors.New("empty answer")
	}

	return answer, nil
}

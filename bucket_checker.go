package main

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

type Resolver interface {
	IsBucket(string) bool
}

// NewS3Resolver initializes a new S3Resolver
func NewS3Resolver(ns string) (*S3Resolver, error) {
	config, err := getConfig(ns)

	if err != nil {
		return nil, err
	}

	return &S3Resolver{
		dnsClient: dns.Client{},
		config:    *config,
	}, nil
}

type S3Resolver struct {
	dnsClient dns.Client
	config    dns.ClientConfig
}

const s3host = "s3.amazonaws.com"

// IsBucket determines whether this prefix is a valid S3 bucket name.
func (s *S3Resolver) IsBucket(name string) bool {
	result, err := s.resolveCNAME(fmt.Sprintf("%s.%s.", name, s3host))

	if err == nil && !strings.Contains(result, "s3-directional") {
		return true
	}

	return false
}

func getConfig(nameserver string) (*dns.ClientConfig, error) {
	if nameserver != "" {
		addr := net.ParseIP(nameserver)
		if addr != nil {
			return &dns.ClientConfig{
				Servers: []string{addr.String()},
				Port:    "53",
			}, nil
		}

		return nil, errors.New("invalid ip addr")
	}

	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")

	if err != nil {
		return nil, errors.New("could not read local resolver config")
	}

	return config, nil
}

func (s *S3Resolver) resolveCNAME(name string) (string, error) {
	msg := dns.Msg{}
	msg.SetQuestion(name, dns.TypeCNAME)

	addr := net.JoinHostPort(s.config.Servers[0], s.config.Port)
	r, _, err := s.dnsClient.Exchange(&msg, addr)

	if err != nil {
		return "", errors.New("probably a timeout")
	}

	var answer = r.Answer

	if len(answer) == 0 {
		return "", errors.New("no resp")
	}

	var answ = answer[0].(*dns.CNAME).Target

	return answ, err
}

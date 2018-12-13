package main

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"strings"
)

type Resolver interface {
	IsBucket(string) bool
}

func NewS3Resolver() *S3Resolver {
	return &S3Resolver{
		dnsClient: dns.Client{},
	}
}

type S3Resolver struct {
	dnsClient dns.Client
}

const s3host = "s3.amazonaws.com"

// IsBucket determines wheter this prefix is a valid S3 bucket name.
func (s *S3Resolver) IsBucket(name string) bool {
	result, err := s.resolveCNAME(fmt.Sprintf("%s.%s.", name, s3host))

	if err == nil && !strings.Contains(result, "s3-directional") {
		return true
	}

	return false
}

func (s *S3Resolver) resolveCNAME(name string) (string, error) {
	msg := dns.Msg{}
	msg.SetQuestion(name, dns.TypeCNAME)

	// TODO: Allow the name server to be set by the user.
	r, _, err := s.dnsClient.Exchange(&msg, "8.8.8.8:53")

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

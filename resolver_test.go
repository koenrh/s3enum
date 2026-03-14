package main

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func S3DNSServer(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)
	m.Authoritative = true

	target := "s3-1-w.amazonaws.com."

	if m.Question[0].Name == "test.s3.amazonaws.com." {
		target = "s3-us-west-2-w.amazonaws.com."
	}

	m.Answer = append(m.Answer, &dns.CNAME{Hdr: dns.RR_Header{
		Name:   m.Question[0].Name,
		Rrtype: dns.TypeCNAME,
		Class:  dns.ClassINET,
		Ttl:    0,
	}, Target: target})

	w.WriteMsg(m)
}

func RunLocalServer(pc net.PacketConn) (*dns.Server, string, chan error, error) {
	server := &dns.Server{
		PacketConn:   pc,
		ReadTimeout:  time.Hour,
		WriteTimeout: time.Hour,
	}

	waitLock := sync.Mutex{}
	waitLock.Lock()
	server.NotifyStartedFunc = waitLock.Unlock

	addr := pc.LocalAddr().String()
	closer := pc

	fin := make(chan error, 1)

	go func() {
		fin <- server.ActivateAndServe()
		closer.Close()
	}()

	waitLock.Lock()
	return server, addr, fin, nil
}

func RunLocalUDPServer(laddr string) (*dns.Server, string, chan error, error) {
	pc, err := net.ListenPacket("udp", laddr)
	if err != nil {
		return nil, "", nil, err
	}

	return RunLocalServer(pc)
}

func TestExistingBucket(t *testing.T) {
	dns.HandleFunc("test.s3.amazonaws.com.", S3DNSServer)
	defer dns.HandleRemove("test.s3.amazonnaws.com.")

	s, addrstr, _, err := RunLocalUDPServer("127.0.0.1:0")

	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	defer s.Shutdown()

	resolver, err := NewDNSResolver(addrstr)
	if err != nil {
		t.Fatalf("unable to create resolver: %v", err)
	}

	if !resolver.IsBucket(context.Background(), "test") {
		t.Fatal("'test' should be an existing bucket")
	}
}

func TestNonExistingBucket(t *testing.T) {
	dns.HandleFunc("testnonexistingbucket.s3.amazonaws.com.", S3DNSServer)
	defer dns.HandleRemove("testnonexistingbucket.s3.amazonnaws.com.")

	s, addrstr, _, err := RunLocalUDPServer("127.0.0.1:0")

	if err != nil {
		t.Fatalf("unable to run test server: %v", err)
	}
	defer s.Shutdown()

	resolver, err := NewDNSResolver(addrstr)
	if err != nil {
		t.Fatalf("unable to create resolver: %v", err)
	}

	if resolver.IsBucket(context.Background(), "testnonexistingbucket") {
		t.Fatal("'testnonexistingbucket' should not be a bucket")
	}
}

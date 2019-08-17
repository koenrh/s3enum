package main

import (
	"fmt"
	"strings"
	"testing"
)

func NewTestResolver() *TestResolver {
	return &TestResolver{}
}

type TestResolver struct{}

func (s *TestResolver) IsBucket(name string) bool {
	return strings.Contains(name, "s3")
}

func TestConsume(t *testing.T) {
	inputChannel := make(chan string)
	resultChannel := make(chan string)
	done2 := make(chan bool)

	done3 := make(chan bool)

	resolver := NewTestResolver()
	consumer := NewConsumer(resolver, inputChannel, resultChannel, done2)

	go consumer.Consume()

	var got []string

	// consumer
	go func() {
		for {
			j, more := <-resultChannel
			if more {
				got = append(got, j)
			} else {
				done3 <- true
				return
			}
		}
	}()

	// producer
	for k := 1; k <= 5; k++ {
		inputChannel <- fmt.Sprintf("test%v", k)
	}
	inputChannel <- "foos3"
	inputChannel <- "foos3asdf"

	close(inputChannel)

	<-done2

	close(resultChannel)
	// block
	<-done3

	// assertions
	expected := []string{"foos3", "foos3asdf"}

	if len(expected) != len(got) {
		t.Fatalf("expected %v, got %v", expected, got)
	}

	for i := range got {
		if got[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	}
}

package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
)

type TestResolver struct{}

func (s *TestResolver) IsBucket(_ context.Context, name string) bool {
	return strings.Contains(name, "s3")
}

func (s *TestResolver) Stats() Stats {
	return Stats{}
}

func TestConsume(t *testing.T) {
	input := make(chan string)
	results := make(chan string)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consume(context.Background(), &TestResolver{}, input, results)
	}()

	var got []string
	var collectWg sync.WaitGroup
	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for j := range results {
			got = append(got, j)
		}
	}()

	for k := 1; k <= 5; k++ {
		input <- fmt.Sprintf("test%v", k)
	}
	input <- "foos3"
	input <- "foos3asdf"

	close(input)
	wg.Wait()
	close(results)
	collectWg.Wait()

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

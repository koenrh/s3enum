package main

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
)

func TestPrintResults(t *testing.T) {
	channel := make(chan string)

	var buf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		printResults(channel, &buf)
	}()

	for i := 1; i <= 5; i++ {
		channel <- fmt.Sprintf("test%v", i)
	}

	close(channel)
	wg.Wait()

	expected := "test1\n" +
		"test2\n" +
		"test3\n" +
		"test4\n" +
		"test5\n"

	if got := buf.String(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPrintResults(t *testing.T) {
	channel := make(chan string)
	done := make(chan bool)

	out := new(bytes.Buffer) // replace 'out' in order to capture the output
	printer := NewPrinter(channel, done, out)

	go printer.PrintBuckets()

	// produce some test results to the results channel
	for i := 1; i <= 5; i++ {
		channel <- fmt.Sprintf("test%v", i)
	}

	close(channel)
	<-done

	expected := "test1\n" +
		"test2\n" +
		"test3\n" +
		"test4\n" +
		"test5\n"

	got := printer.out.(*bytes.Buffer).String()
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

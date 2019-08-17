package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPrintResults(t *testing.T) {
	channel := make(chan string)
	done := make(chan bool)

	printer := NewPrinter(channel, done)

	go printer.PrintBuckets()

	out = new(bytes.Buffer) // replace 'out' in order to capture the output

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

	got := out.(*bytes.Buffer).String()
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

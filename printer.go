package main

import (
	"fmt"
	"io"
	"os"
)

var (
	out io.Writer = os.Stdout // substituted during testing
)

// Printer struct
type Printer struct {
	channel chan string
	done    chan bool
}

// NewPrinter initializer
func NewPrinter(channel chan string, done chan bool) (*Printer, error) {
	printer := &Printer{
		channel: channel,
		done:    done,
	}

	return printer, nil
}

// PrintBuckets prints the results as they come in.
func (c *Printer) PrintBuckets() {
	for {
		bucket, more := <-c.channel
		if more {
			fmt.Fprintf(out, "%s\n", bucket)
		} else {
			c.done <- true
			return
		}
	}
}

package main

import (
	"fmt"
	"io"
)

// Printer struct
type Printer struct {
	channel chan string
	done    chan bool
	out     io.Writer
}

// NewPrinter initializer
func NewPrinter(channel chan string, done chan bool, out io.Writer) *Printer {
	printer := &Printer{
		channel: channel,
		done:    done,
		out:     out,
	}

	return printer
}

// PrintBuckets prints the results as they come in.
func (c *Printer) PrintBuckets() {
	for {
		bucket, more := <-c.channel
		if more {
			fmt.Fprintf(c.out, "%s\n", bucket)
		} else {
			c.done <- true
			return
		}
	}
}

package main

import (
	"fmt"
	"io"
)

// Printer struct
type Printer struct {
	channel chan string
	done    chan bool
	log     io.Writer
}

// NewPrinter initializer
func NewPrinter(channel chan string, done chan bool, log io.Writer) *Printer {
	printer := &Printer{
		channel: channel,
		done:    done,
		log:     log,
	}

	return printer
}

// PrintBuckets prints the results as they come in.
func (c *Printer) PrintBuckets() {
	for {
		bucket, more := <-c.channel
		if more {
			fmt.Fprintf(c.log, "%s\n", bucket)
		} else {
			c.done <- true
			return
		}
	}
}

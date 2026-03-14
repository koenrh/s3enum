package main

import (
	"fmt"
	"io"
)

func printResults(results <-chan string, w io.Writer) {
	for bucket := range results {
		fmt.Fprintln(w, bucket)
	}
}

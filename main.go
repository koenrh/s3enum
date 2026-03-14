package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"time"
)

const version = "1.1.0"

func main() {
	wordListPtr := flag.String("wordlist", "", "Path to word list")
	suffixListPtr := flag.String("suffixlist", "", "Path to suffix list")
	threadsPtr := flag.Int("threads", 50, "Number of concurrent workers")
	nameServerPtr := flag.String("nameserver", "", "Custom name server")
	versionPtr := flag.Bool("version", false, "Print version")

	flag.Parse()

	if *versionPtr {
		fmt.Println("v" + version)
		return
	}

	names := flag.Args()

	if *suffixListPtr == "" || *wordListPtr == "" || len(names) == 0 {
		fmt.Println("s3enum -wordlist wordlist.txt -suffixlist suffixlistt.txt [-threads 5] [-nameserver 1.1.1.1] <name>...")
		flag.PrintDefaults()
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	wordChannel := make(chan string, 1000)
	resultChannel := make(chan string)

	resolver, err := NewDNSResolver(*nameServerPtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not initialize DNS resolver: %v\n", err)
		os.Exit(1)
	}

	start := time.Now()

	var workerWg sync.WaitGroup
	for i := 0; i < *threadsPtr; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			consume(ctx, resolver, wordChannel, resultChannel)
		}()
	}

	var printerWg sync.WaitGroup
	printerWg.Add(1)
	go func() {
		defer printerWg.Done()
		printResults(resultChannel, os.Stdout)
	}()

	producer, err := NewProducer(*suffixListPtr, wordChannel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not initialize producer: %v\n", err)
		os.Exit(1)
	}

	if err := producer.ProduceWordList(ctx, names, *wordListPtr); err != nil {
		fmt.Fprintf(os.Stderr, "error producing word list: %v\n", err)
	}

	workerWg.Wait()
	close(resultChannel)
	printerWg.Wait()

	stats := resolver.Stats()
	stats.Duration = time.Since(start)
	fmt.Fprintf(os.Stderr, "\n%s\n", stats.Summary())
}

func consume(ctx context.Context, resolver Resolver, input <-chan string, results chan<- string) {
	for name := range input {
		if ctx.Err() != nil {
			return
		}
		if resolver.IsBucket(ctx, name) {
			results <- name
		}
	}
}

func printResults(results <-chan string, w io.Writer) {
	for bucket := range results {
		fmt.Fprintln(w, bucket)
	}
}

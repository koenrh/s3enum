package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	threads            int
	baseName           string
	wordListFile       string
	preAndSuffixesFile string
)

func main() {
	flag.StringVar(&baseName, "n", "", "base name")
	flag.StringVar(&preAndSuffixesFile, "p", "", "path to the prefixes file")
	flag.StringVar(&wordListFile, "w", "", "path to the word list")
	flag.IntVar(&threads, "t", 10, "number of threads")
	flag.Parse()

	wordChannel := make(chan string)
	wordDone := make(chan bool)

	resultChannel := make(chan string)
	resultDone := make(chan bool)

	resolver := NewS3Resolver()

	consumer, err := NewConsumer(resolver, wordChannel, resultChannel, wordDone)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize Consumer: %v\n", err)
		os.Exit(1)
	}

	for i := 0; i < threads; i++ {
		go consumer.Consume()
	}

	printer, err := NewPrinter(resultChannel, resultDone)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize Printer: %v\n", err)
		os.Exit(1)
	}
	go printer.PrintBuckets()

	producer, err := NewProducer(preAndSuffixesFile, wordChannel, resultDone)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize Producer: %v\n", err)
		os.Exit(1)
	}

	producer.ProduceWordList(baseName, wordListFile)

	// NOTE: producer closes their own channel
	<-wordDone

	close(resultChannel)
	<-resultDone
}

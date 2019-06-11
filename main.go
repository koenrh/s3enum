package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"os"
)

var (
	threads            int
	names              []string
	wordListFile       string
	preAndSuffixesFile string
)

const version = "0.0.1"
const usage = `s3enum

Usage:
  s3enum --wordlist wl.txt --suffixlist sl.txt [--threads 2] <name>...
  s3enum -h | --help
  s3enum --version

Options:
  --wordlist <path>    Path to the word list.
  --suffixlist <path>  Path to the word list.
  --threads <threads>  Number of threads [default: 10].
  -h --help            Show this screen.`

func main() {
	opts, err := docopt.ParseDoc(usage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Bad arguments")
		os.Exit(1)
	}
	if opts["--version"].(bool) {
		fmt.Println(version)
		os.Exit(0)
	}

	names = opts["<name>"].([]string)
	preAndSuffixesFile = opts["--suffixlist"].(string)
	wordListFile = opts["--wordlist"].(string)
	threads, _ = opts.Int("--threads")

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

	producer.ProduceWordList(names, wordListFile)

	// NOTE: producer closes their own channel
	<-wordDone

	close(resultChannel)
	<-resultDone
}

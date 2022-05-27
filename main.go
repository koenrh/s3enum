package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.2.0"

func main() {
	wordListPtr := flag.String("wordlist", "", "Path to word list")
	suffixListPtr := flag.String("suffixlist", "", "Path to suffix list")
	threadsPtr := flag.Int("threads", 5, "Number of threads")
	nameServerPtr := flag.String("nameserver", "", "Custom name server")
	versionPtr := flag.Bool("version", false, "Print version")

	flag.Parse()

	if *versionPtr {
		fmt.Println("v" + version)
		return
	}

	var names = flag.Args()

	if *suffixListPtr == "" || *wordListPtr == "" || len(names) == 0 {
		fmt.Println("s3enum -wordlist wordlist.txt -suffixlist suffixlistt.txt [-threads 5] [-nameserver 1.1.1.1] <name>...")
		flag.PrintDefaults()
		os.Exit(1)
	}

	wordChannel := make(chan string)
	wordDone := make(chan bool)

	resultChannel := make(chan string)
	resultDone := make(chan bool)

	resolver, err := NewDNSResolver(*nameServerPtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize DNS resolver: %v\n", err)
		os.Exit(1)
	}

	consumer := NewConsumer(resolver, wordChannel, resultChannel, wordDone)

	for i := 0; i < *threadsPtr; i++ {
		go consumer.Consume()
	}

	printer := NewPrinter(resultChannel, resultDone, os.Stdout)
	go printer.PrintBuckets()

	producer, err := NewProducer(*suffixListPtr, wordChannel, resultDone)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize Producer: %v\n", err)
		os.Exit(1)
	}

	producer.ProduceWordList(names, *wordListPtr)

	// NOTE: producer closes their own channel
	<-wordDone

	close(resultChannel)
	<-resultDone
}

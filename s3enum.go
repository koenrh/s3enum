package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"os"
	"strings"
	"sync"
)

const (
	s3host = "s3.amazonaws.com"
)

var (
	suffixes = [...]string{
		"approval",
		"dev",
		"development",
		"live",
		"stag",
		"staging",
		"prod",
		"production",
		"test",
	}
)

// CheckName checks whether a bucket exists for (variations of) the name.
func CheckName(name string, word string, resultsChannel chan<- string) {

	// TODO: extend with pre- and suffixes
	candidates := []string{
		fmt.Sprintf("%s-%s", name, word),
		fmt.Sprintf("%s-%s", word, name),
		fmt.Sprintf("%s_%s", name, word),
		fmt.Sprintf("%s_%s", word, name),
		fmt.Sprintf("%s.%s", name, word),
		fmt.Sprintf("%s.%s", word, name),
		fmt.Sprintf("%s%s", name, word),
		fmt.Sprintf("%s%s", word, name),
	}
	for _, suf := range suffixes {
		reg := []string{name, word, suf}
		candidates = append(candidates, strings.Join(reg[:], "-"))
		candidates = append(candidates, strings.Join(reg[:], "_"))
		candidates = append(candidates, strings.Join(reg[:], "."))
		candidates = append(candidates, strings.Join(reg[:], ""))
	}

	for _, candidate := range candidates {
		result := resolveCNAME(fmt.Sprintf("%s.%s.", candidate, s3host))
		if len(result) == 0 {
			// No results
			continue
		}

		if v, ok := result[0].(*dns.CNAME); ok {
			if !strings.Contains(v.Target, "s3-directional") {
				resultsChannel <- candidate
			}
		}
	}
}

var (
	threads  int
	baseName string
	wordList string
)

func main() {
	flag.StringVar(&baseName, "n", "", "base name")
	flag.StringVar(&wordList, "w", "", "path to the word list")
	flag.IntVar(&threads, "t", 10, "number of threads")
	flag.Parse()

	wordsChannel := make(chan string)
	resultsChannel := make(chan string)

	producerGroup := new(sync.WaitGroup)
	producerGroup.Add(threads)

	consumerGroup := new(sync.WaitGroup)
	consumerGroup.Add(1)

	// Create a goroutine for the number of threads specified.
	for i := 0; i < threads; i++ {
		go func() {
			for {
				word := <-wordsChannel

				// Check if we need to continue.
				if word == "" {
					break
				}

				CheckName(baseName, word, resultsChannel)
			}

			// Signal to the wait gropu that the thread has finished.
			producerGroup.Done()
		}()
	}

	// Consumer goroutine that prints the results as they appear.
	go func() {
		for r := range resultsChannel {
			fmt.Println(r)
		}
		consumerGroup.Done()
	}()

	// Read word list
	file, err := os.Open(wordList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		wordsChannel <- line
	}

	close(wordsChannel)
	producerGroup.Wait()
	close(resultsChannel)
	consumerGroup.Wait()
}

func resolveCNAME(name string) []dns.RR {
	msg := dns.Msg{}
	msg.SetQuestion(name, dns.TypeCNAME)

	// TODO: Set/reduce timeout? (ReadTimeout)
	client := &dns.Client{Net: "tcp"}

	// TODO: Allow the name server to be set by the user.
	r, _, err := client.Exchange(&msg, "8.8.8.8:53")

	if err != nil {
		return nil
	}

	return r.Answer
}

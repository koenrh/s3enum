package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

// Producer struct
type Producer struct {
	channel        chan string
	quit           chan bool
	delimiters     []string
	preAndSuffixes []string
}

// NewProducer initializer
func NewProducer(preAndSuffixesFile string, wordChannel chan string, quit chan bool) (*Producer, error) {
	producer := &Producer{
		channel:    wordChannel,
		quit:       quit,
		delimiters: []string{"-", "_", ".", ""},
	}

	contents, err := readFile(preAndSuffixesFile)
	if err != nil {
		return nil, errors.New("failed to read pre- and suffxies file")
	}
	producer.preAndSuffixes = contents

	return producer, nil
}

// ProduceWordList produces candidate bucket names to the channel.
func (p *Producer) ProduceWordList(basename string, list string) {
	p.channel <- basename

	file, err := os.Open(list)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		p.Produce(basename, line)
	}

	close(p.channel)
}

// Produce produces candidates
func (p *Producer) Produce(basename string, word string) {
	for _, ca := range p.PrepareCandidateBucketNames(basename, word) {
		p.channel <- ca
	}
}

// PrepareCandidateBucketNames creates all candidate pairs
func (p *Producer) PrepareCandidateBucketNames(basename string, word string) []string {
	result := []string{}

	for _, del := range p.delimiters {
		cand1 := fmt.Sprintf("%s%s%s", basename, del, word)
		cand2 := fmt.Sprintf("%s%s%s", word, del, basename)

		result = append(result, cand1)
		result = append(result, cand2)

		for _, ca := range p.preAndSuffixes {
			result = append(result, fmt.Sprintf("%s%s%s", cand1, del, ca))
			result = append(result, fmt.Sprintf("%s%s%s", ca, del, cand1))

			result = append(result, fmt.Sprintf("%s%s%s", cand2, del, ca))
			result = append(result, fmt.Sprintf("%s%s%s", ca, del, cand2))
		}
	}

	return result
}

func readFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

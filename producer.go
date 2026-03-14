package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
)

type Producer struct {
	channel        chan string
	delimiters     []string
	preAndSuffixes []string
}

func NewProducer(preAndSuffixesFile string, wordChannel chan string) (*Producer, error) {
	contents, err := readFile(preAndSuffixesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read suffixes file: %w", err)
	}

	return &Producer{
		channel:        wordChannel,
		delimiters:     []string{"-", "_", ".", ""},
		preAndSuffixes: contents,
	}, nil
}

func (p *Producer) ProduceWordList(ctx context.Context, names []string, list string) error {
	defer close(p.channel)

	for _, n := range names {
		p.channel <- n
	}

	file, err := os.Open(list)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := scanner.Text()
		for _, n := range names {
			p.Produce(n, line)
		}
	}

	return scanner.Err()
}

func (p *Producer) Produce(name, word string) {
	for _, del := range p.delimiters {
		cand1 := name + del + word
		cand2 := word + del + name

		p.channel <- cand1
		p.channel <- cand2

		for _, affix := range p.preAndSuffixes {
			p.channel <- cand1 + del + affix
			p.channel <- affix + del + cand1
			p.channel <- cand2 + del + affix
			p.channel <- affix + del + cand2
		}
	}
}

// PrepareCandidateBucketNames creates all candidate pairs.
func (p *Producer) PrepareCandidateBucketNames(name, word string) []string {
	perDelimiter := 2 + 4*len(p.preAndSuffixes)
	result := make([]string, 0, len(p.delimiters)*perDelimiter)

	for _, del := range p.delimiters {
		cand1 := name + del + word
		cand2 := word + del + name

		result = append(result, cand1)
		result = append(result, cand2)

		for _, affix := range p.preAndSuffixes {
			result = append(result, cand1+del+affix)
			result = append(result, affix+del+cand1)
			result = append(result, cand2+del+affix)
			result = append(result, affix+del+cand2)
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

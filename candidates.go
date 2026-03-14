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
		if !p.send(ctx, n) {
			return ctx.Err()
		}
	}

	file, err := os.Open(list)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		for _, n := range names {
			if !p.Produce(ctx, n, line) {
				return ctx.Err()
			}
		}
	}

	return scanner.Err()
}

// Produce generates all candidate bucket names for a given name and word
// combination. Returns false if the context was cancelled.
func (p *Producer) Produce(ctx context.Context, name, word string) bool {
	for _, del := range p.delimiters {
		cand1 := name + del + word
		cand2 := word + del + name

		if !p.send(ctx, cand1) || !p.send(ctx, cand2) {
			return false
		}

		for _, affix := range p.preAndSuffixes {
			if !p.send(ctx, cand1+del+affix) ||
				!p.send(ctx, affix+del+cand1) ||
				!p.send(ctx, cand2+del+affix) ||
				!p.send(ctx, affix+del+cand2) {
				return false
			}
		}
	}
	return true
}

func (p *Producer) send(ctx context.Context, candidate string) bool {
	select {
	case p.channel <- candidate:
		return true
	case <-ctx.Done():
		return false
	}
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

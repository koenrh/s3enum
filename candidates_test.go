package main

import (
	"sync"
	"testing"
)

func TestProduceCandidateNamesToChannel(t *testing.T) {
	channel := make(chan string)

	producer, err := NewProducer("testdata/suffixlist.txt", channel)
	if err != nil {
		t.Fatalf("could not initialize Producer: %v", err)
	}

	producer.delimiters = []string{"-"}

	var got []string

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := range channel {
			got = append(got, j)
		}
	}()

	producer.Produce("foo", "bar")
	close(channel)
	wg.Wait()

	expected := []string{
		"foo-bar",
		"bar-foo",
		"foo-bar-baz",
		"baz-foo-bar",
		"bar-foo-baz",
		"baz-bar-foo",
	}

	if len(expected) != len(got) {
		t.Fatalf("expected %v, got %v", expected, got)
	}

	for i := range got {
		if got[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	}
}

func TestFindCandidates(t *testing.T) {
	channel := make(chan string)

	producer, err := NewProducer("testdata/suffixlist.txt", channel)
	if err != nil {
		t.Fatalf("could not initialize Producer: %v", err)
	}

	producer.delimiters = []string{"."}

	expected := []string{
		"x.y",
		"y.x",
		"x.y.baz",
		"baz.x.y",
		"y.x.baz",
		"baz.y.x",
	}

	got := producer.PrepareCandidateBucketNames("x", "y")

	if len(expected) != len(got) {
		t.Fatalf("expected %v, got %v", expected, got)
	}

	for i := range got {
		if got[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	}
}

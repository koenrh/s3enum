package main

import (
	"testing"
)

func TestProduceCandidateNamesToChannel(t *testing.T) {
	channel := make(chan string)
	done := make(chan bool)

	producer, err := NewProducer("testdata/suffixlist.txt", channel, done)
	if err != nil {
		t.Errorf("could not initialize Producer")
	}

	producer.delimiters = []string{"-"}

	var got []string

	// consumer
	go func() {
		for {
			j, more := <-channel
			if more {
				got = append(got, j)
			} else {
				done <- true
				return
			}
		}
	}()

	producer.Produce("foo", "bar")
	close(channel)

	<-done

	// assertions
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
	done := make(chan bool)

	producer, err := NewProducer("testdata/suffixlist.txt", channel, done)
	if err != nil {
		t.Errorf("could not initialize Producer")
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

package main

import "context"

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

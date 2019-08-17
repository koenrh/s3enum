package main

// Consumer struct
type Consumer struct {
	inputChannel  chan string
	resultChannel chan string
	quit          chan bool
	resolver      Resolver
}

// NewConsumer initializer
func NewConsumer(resolver Resolver, input chan string, result chan string, quit chan bool) *Consumer {
	consumer := &Consumer{
		resolver:      resolver,
		inputChannel:  input,
		resultChannel: result,
		quit:          quit,
	}

	return consumer
}

// Consume reads messages from 'input', and outputs results to 'result'.
func (c *Consumer) Consume() {
	for {
		j, more := <-c.inputChannel
		if more {
			if c.resolver.IsBucket(j) {
				c.resultChannel <- j
			}
		} else {
			c.quit <- true
			return
		}
	}
}

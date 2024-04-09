package queue

import (
	"estimation-poker/internal/domain"
	"log"
)

type ChanQueue struct {
	ch chan domain.Event
}

func NewChanQueue() *ChanQueue {
	return &ChanQueue{make(chan domain.Event, 1)}
}

func (c *ChanQueue) Publish(event domain.Event) error {
	log.Printf("%#v\n", event)
	c.ch <- event
	return nil
}

func (c *ChanQueue) Subscribe() (<-chan domain.Event, error) {
	return c.ch, nil
}

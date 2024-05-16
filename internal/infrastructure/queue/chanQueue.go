package queue

import (
	"estimation-poker/internal/domain"
)

type ChanQueue struct {
	ch chan domain.Event
}

func NewChanQueue() *ChanQueue {
	return &ChanQueue{make(chan domain.Event, 1)}
}

func (c *ChanQueue) Publish(event domain.Event) error {
	c.ch <- event
	return nil
}

func (c *ChanQueue) Subscribe() (<-chan domain.Event, error) {
	return c.ch, nil
}

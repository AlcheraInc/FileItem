package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"errors"
)

type rupdate struct {
	item  Item
	fails chan<- error
}
type rcreate struct {
	item2
	fails chan<- error
}
type rdelete struct {
	rcreate
}
type rsearch struct {
	rcreate
}

type itemAgent struct {
	exits, grace chan error
	searches     chan rsearch
	creates      chan rcreate
	updates      chan rupdate
	deletes      chan rdelete
}

func (h *itemAgent) setup(backlog int) {
	h.exits, h.grace = make(chan error, 1), make(chan error, 1)
	h.searches = make(chan rsearch, backlog)
	h.creates = make(chan rcreate, backlog)
	h.updates = make(chan rupdate, backlog)
	h.deletes = make(chan rdelete, backlog)
}

func (h *itemAgent) stop(err error) {
	// signal to stop.
	// we won't care about the instance
	h.exits <- err
	for range h.grace {
		// wait until it's closed
	}
}

type itemCallbacks interface {
	onSearchItem(msg rsearch)
	onCreateItem(msg rcreate)
	onUpdateItem(msg rupdate)
	onDeleteItem(msg rdelete)
}

// handle leftover requests
func (h *itemAgent) cleanup(callbacks itemCallbacks) {
	for {
		if err := h.peek(callbacks); err != nil {
			return
		}
	}
}

func (h *itemAgent) peek(callbacks itemCallbacks) error {
	select {
	case msg := <-h.creates:
		callbacks.onCreateItem(msg)
	case msg := <-h.deletes:
		callbacks.onDeleteItem(msg)
	case msg := <-h.updates:
		callbacks.onUpdateItem(msg)
	case msg := <-h.searches:
		callbacks.onSearchItem(msg)
	default:
		return errors.New("no more requests in the queue")
	}
	return nil
}

func (h *itemAgent) start(callbacks itemCallbacks) {
	defer close(h.grace)
	defer close(h.exits)
	defer h.cleanup(callbacks) // loop on destruction

	for { // loop until exit
		select {
		case <-h.exits:
			return
		case msg := <-h.creates:
			callbacks.onCreateItem(msg)
		case msg := <-h.deletes:
			callbacks.onDeleteItem(msg)
		case msg := <-h.updates:
			callbacks.onUpdateItem(msg)
		case msg := <-h.searches:
			callbacks.onSearchItem(msg)
		}
	}
}

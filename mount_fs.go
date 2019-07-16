package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"errors"
	"log"
	"os"
)

type faccess struct {
	fpath string
	foper FileOp
	perm  os.FileMode
	fails chan<- error
}

type fsMount struct {
	exits, grace chan error
	accesses     chan faccess
}

func (h *fsMount) setup(backlog int) {
	h.exits, h.grace = make(chan error, 1), make(chan error, 1)
	h.accesses = make(chan faccess, backlog)
}

func (h *fsMount) stop(err error) {
	// signal to stop.
	// we won't care about the instance
	h.exits <- err
	for range h.grace {
		// wait until it's closed
	}
}

type fileCallbacks interface {
	inUse(fpath string) bool
	remember(fpath string)
	forget(fpath string)
}

func (h *fsMount) record(stat os.FileInfo) {
	// for now, do noting but logging
	log.Println(stat.Name(), stat.Size())
}

func (h *fsMount) serveOne(a faccess, r fileCallbacks) {
	defer r.forget(a.fpath)
	r.remember(a.fpath)

	defer close(a.fails)
	stat, err := os.Stat(a.fpath)
	if err != nil {
		a.fails <- err
		return
	}
	h.record(stat)

	if err := a.foper(os.OpenFile(a.fpath, os.O_RDWR|os.O_SYNC, a.perm)); err != nil {
		a.fails <- err
	}
}

func (h *fsMount) pend(a faccess) {
	h.accesses <- a
}

func (h *fsMount) peek(callbacks fileCallbacks) error {
	select {
	case a := <-h.accesses:
		if callbacks.inUse(a.fpath) {
			go h.pend(a)
		}
		h.serveOne(a, callbacks)
	default:
		return errors.New("no more requests in the queue")
	}
	return nil
}

// handle leftover requests
func (h *fsMount) cleanup(callbacks fileCallbacks) {
	for {
		if err := h.peek(callbacks); err != nil {
			return
		}
	}
}

func (h *fsMount) start(callbacks fileCallbacks) {
	defer close(h.grace)
	defer close(h.exits)
	defer h.cleanup(callbacks) // loop on destruction

	for { // loop until exit
		select {
		case <-h.exits:
			return
		case a := <-h.accesses:
			if callbacks.inUse(a.fpath) {
				go h.pend(a)
			}
			h.serveOne(a, callbacks)
		}
	}
}

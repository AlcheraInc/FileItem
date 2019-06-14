package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"io"
	"os"
)

type rcreate struct {
	iname, itype string
	results      chan<- FileGroupItem
}
type rdelete struct {
	iname, itype string
	fails        chan<- error
}
type rsearch struct {
	iname, itype string
	results      chan<- FileGroupItem
}
type rupdate struct {
	item   FileGroupItem
	detail map[string]interface{}
	fails  chan<- error
}

type rsave struct {
	fpath    string
	contents io.Reader
	fails    chan<- error
}
type rremove struct {
	fpath string
	fails chan<- error
}

type rload struct {
	fpath   string
	results chan<- *os.File
}

type receiver interface {
	onSearchType(itype string, results chan<- FileGroupItem)
	onSearch(iname, itype string, results chan<- FileGroupItem, doClose bool)
	onCreateItem(msg *rcreate)
	onDeleteItem(msg *rdelete)
	onUpdateItem(msg *rupdate)
	onSaveFile(msg *rsave)
	onRemoveFile(msg *rremove)
	onLoadFile(msg *rload)
}

type rhub struct {
	exits, grace chan error

	searches chan *rsearch

	creates chan *rcreate
	deletes chan *rdelete
	updates chan *rupdate

	saves   chan *rsave
	loads   chan *rload
	removes chan *rremove
}

func (h *rhub) setup() {
	const capacity = 10

	h.exits, h.grace = make(chan error, 1), make(chan error, 1)
	h.saves = make(chan *rsave, capacity)
	h.loads = make(chan *rload, capacity)
	h.searches = make(chan *rsearch, capacity)
	h.creates = make(chan *rcreate, capacity)
	h.deletes = make(chan *rdelete, capacity)
	h.updates = make(chan *rupdate, capacity)
	h.removes = make(chan *rremove, capacity)
}

func (h *rhub) stop(err error) {
	h.exits <- err
}

func (h *rhub) loop(r receiver) {
	defer close(h.grace)
	defer close(h.exits)
	for {
		select {
		case msg := <-h.creates:
			r.onCreateItem(msg)
		case msg := <-h.deletes:
			r.onDeleteItem(msg)
		case msg := <-h.updates:
			r.onUpdateItem(msg)

		case msg := <-h.saves:
			r.onSaveFile(msg)
		case msg := <-h.removes:
			r.onRemoveFile(msg)
		case msg := <-h.loads:
			r.onLoadFile(msg)

		case msg := <-h.searches:
			if len(msg.iname) == 0 {
				r.onSearchType(msg.itype, msg.results)
			} else {
				r.onSearch(msg.iname, msg.itype, msg.results, true)
			}
		case <-h.exits:
			return
		}
	}
}

package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import "io"

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
type rload struct {
	fpath   string
	results chan<- io.ReadWriteCloser
}

type receiver interface {
	onSearchType(itype string, results chan<- FileGroupItem)
	onSearch(iname, itype string, results chan<- FileGroupItem, doClose bool)
	onCreateItem(msg *rcreate)
	onDeleteItem(msg *rdelete)
	onUpdateItem(msg *rupdate)
	onSaveFile(msg *rsave)
	onLoadFile(msg *rload)
}

type rhub struct {
	exits, grace chan error
	saves        chan *rsave
	loads        chan *rload
	searches     chan *rsearch
	creates      chan *rcreate
	deletes      chan *rdelete
	updates      chan *rupdate
}

func (h *rhub) setup() {
	h.exits, h.grace = make(chan error, 1), make(chan error, 1)
	h.saves = make(chan *rsave)
	h.loads = make(chan *rload)
	h.searches = make(chan *rsearch)
	h.creates = make(chan *rcreate)
	h.deletes = make(chan *rdelete)
	h.updates = make(chan *rupdate)
}

func (h *rhub) stop(err error) {
	h.exits <- err
}

func (h *rhub) loop(r receiver) {
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
		case msg := <-h.loads:
			r.onLoadFile(msg)
		case msg := <-h.searches:
			if len(msg.iname) == 0 {
				r.onSearchType(msg.itype, msg.results)
			} else {
				r.onSearch(msg.iname, msg.itype, msg.results, true)
			}
		case err := <-h.exits:
			h.grace <- err
			return
		}
	}
}

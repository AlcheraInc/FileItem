package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const cacheFile, cacheFormat = "info.json", "json"

type fileset1 struct {
	name, folder string
	rhub
}

func NewOwner1(name string) (ItemOwner, error) {
	wd, _ := os.Getwd()

	set := new(fileset1)
	set.rhub.setup()
	set.name, set.folder = name, filepath.Join(wd, name)

	os.Mkdir(set.folder, 0775)
	go set.rhub.loop(set)
	return set, nil
}

func (set *fileset1) GetPath() string {
	return set.folder
}
func (set *fileset1) Close() error {
	set.stop(nil)
	return nil
}
func (set *fileset1) Remove() {
	os.RemoveAll(set.folder)
	set.Close()
	<-set.grace
}

func (set *fileset1) NewItem(iname, itype string, detail map[string]interface{}) FileGroupItem {
	ch := make(chan FileGroupItem)
	go set.create(iname, itype, ch)

	item, errs := <-ch, make(chan error, 1)
	go set.update(item, detail, errs)
	for range errs {
		log.Fatalln("update error")
		return nil
	}
	return item
}
func (set *fileset1) create(iname, itype string, ch chan<- FileGroupItem) {
	req := new(rcreate)
	req.iname, req.itype = iname, itype
	req.results = ch
	set.creates <- req
}
func (set *fileset1) onCreateItem(msg *rcreate) {
	defer close(msg.results)

	tpath := filepath.Join(set.folder, msg.itype)
	os.Mkdir(tpath, 0775)
	ipath := filepath.Join(tpath, msg.iname)
	os.Mkdir(ipath, 0775)

	item := newitem1(msg.iname, msg.itype)
	item.Ctime = time.Now()
	item.root = set.folder
	item.owner = set
	msg.results <- item
}

func (set *fileset1) Delete(item Item) <-chan error {
	errs := make(chan error, 1)
	if item == nil {
		close(errs)
		return errs
	}
	go set.delete(item.GetName(), item.GetType(), errs)
	return errs
}
func (set *fileset1) delete(iname, itype string, errs chan<- error) {
	req := new(rdelete)
	req.iname, req.itype = iname, itype
	req.fails = errs
	set.deletes <- req
}
func (set *fileset1) onDeleteItem(msg *rdelete) {
	defer close(msg.fails)
	ipath := filepath.Join(set.folder, msg.itype, msg.iname)

	if err := os.RemoveAll(ipath); err != nil {
		msg.fails <- err
	}
}

func (set *fileset1) FindNames(itype string) <-chan string {
	items, names := make(chan FileGroupItem, 5), make(chan string)
	go set.search("", itype, items)
	go func() {
		defer close(names)
		for item := range items {
			names <- item.GetName()
		}
	}()
	return names
}
func (set *fileset1) FindOne(hint Item, receiver interface{}) <-chan error {
	errs := make(chan error, 1)
	if receiver == nil {
		errs <- errors.New("null argument: receiver")
		close(errs)
		return errs
	}
	items := make(chan FileGroupItem, 1)
	go set.search(hint.GetName(), hint.GetType(), items)
	go func() {
		defer close(errs)
		item := <-items
		if item == nil {
			return
		}
		file := <-item.LoadFile(cacheFile, cacheFormat)
		if file == nil {
			errs <- fmt.Errorf("failed to load '%s'", cacheFile)
			return
		}
		dec := json.NewDecoder(file)
		if err := dec.Decode(receiver); err != nil {
			errs <- err
			return
		}
	}()
	return errs
}
func (set *fileset1) search(iname, itype string, ch chan<- FileGroupItem) {
	req := new(rsearch)
	req.iname, req.itype = iname, itype
	req.results = ch
	set.searches <- req
}
func (set *fileset1) onSearchType(itype string, results chan<- FileGroupItem) {
	defer close(results)
	files, err := ioutil.ReadDir(filepath.Join(set.folder, itype))
	if err != nil {
		return
	}
	for _, f := range files {
		if f.IsDir() == false {
			continue
		}
		set.onSearch(f.Name(), itype, results, false)
	}
}
func (set *fileset1) onSearch(iname, itype string, results chan<- FileGroupItem, doClose bool) {
	if doClose {
		defer close(results)
	}
	item := newitem1(iname, itype)
	item.root = set.folder
	item.owner = set
	results <- item
}

func (set *fileset1) Update(hint Item, changes map[string]interface{}) <-chan error {
	items, errs := make(chan FileGroupItem), make(chan error)
	go set.search(hint.GetName(), hint.GetType(), items)
	if item := <-items; item != nil {
		go set.update(item, changes, errs)
	} else {
		close(errs)
	}
	return errs
}
func (set *fileset1) update(item FileGroupItem, detail map[string]interface{}, errs chan<- error) {
	req := new(rupdate)
	req.item, req.detail = item, detail
	req.fails = errs
	set.updates <- req
}
func (set *fileset1) onUpdateItem(msg *rupdate) {
	item, casted := msg.item.(*item1)
	if casted == false {
		log.Fatalln("type cast failed")
	}
	go func() {
		defer close(msg.fails)
		if msg.detail == nil {
			msg.detail = item.toMap()
		} else {
			item.mergeTo(msg.detail)
		}
		buf, _ := json.Marshal(msg.detail)
		for err := range item.SaveFile(cacheFile, cacheFormat, bytes.NewBuffer(buf)) {
			msg.fails <- err
		}
	}()
}

func (set *fileset1) onSaveFile(msg *rsave) {
	defer close(msg.fails)
	file, err := os.OpenFile(msg.fpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY|os.O_SYNC, 0664)
	if err != nil {
		msg.fails <- err
		return
	}
	defer file.Close()
	if _, err := io.Copy(file, msg.contents); err != nil {
		msg.fails <- err
		return
	}
}
func (set *fileset1) onLoadFile(msg *rload) {
	defer close(msg.results)
	file, err := os.Open(msg.fpath)
	if err != nil {
		return
	}
	msg.results <- file
}

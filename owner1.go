package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const CacheFile, cacheFormat = "info.json", "json"

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
	for range set.grace {
	}
	os.RemoveAll(set.folder)
}

func (set *fileset1) NewItem(iname, itype string, detail map[string]interface{}) <-chan FileGroupItem {
	ch := make(chan FileGroupItem)
	// prevent bad creation with empty name and empty type
	if len(iname) == 0 || len(itype) == 0 {
		close(ch)
		return ch
	}
	creates := make(chan FileGroupItem)
	go set.create(iname, itype, creates)
	go func() {
		defer close(ch)
		for item := range creates {
			for err := range set.Update(item, detail) {
				log.Println(err)
				return
			}
			ch <- item
		}
	}()
	return ch
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

	fpath := filepath.Join(item.GetPath(), CacheFile)
	file, err := os.Create(fpath)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(item); err != nil {
		log.Println(err)
		return
	}
	msg.results <- item
}

func (set *fileset1) Delete(item Item) <-chan error {
	errs := make(chan error)
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

func (set *fileset1) Find(iname, itype string) <-chan FileGroupItem {
	items := make(chan FileGroupItem)
	go set.search(iname, itype, items)
	return items
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

	fpath := filepath.Join(item.GetPath(), CacheFile)
	if _, err := os.Stat(fpath); err != nil {
		log.Println(err)
		return
	}
	file, err := os.Open(fpath)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	if err := enc.Encode(&item); err != nil {
		log.Println(err)
		return
	}
	item.owner = set
	results <- item
}

func (set *fileset1) Update(hint Item, changes map[string]interface{}) <-chan error {
	errs := make(chan error)
	items := make(chan FileGroupItem)
	go set.search(hint.GetName(), hint.GetType(), items)
	go func() {
		set.update(<-items, changes, errs)
	}()
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
		msg.fails <- errors.New("type cast failed")
		close(msg.fails)
		return
	}
	go func() {
		defer close(msg.fails)
		if msg.detail == nil {
			msg.detail = item.toMap()
		} else {
			item.mergeTo(msg.detail)
		}
		buf, _ := json.Marshal(msg.detail)
		for err := range item.SaveFile(CacheFile, cacheFormat, bytes.NewBuffer(buf)) {
			log.Println(err)
			msg.fails <- err
		}
	}()
}

func (set *fileset1) onSaveFile(msg *rsave) {
	file, err := os.OpenFile(msg.fpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY|os.O_SYNC, 0664)
	if err != nil {
		msg.fails <- err
		close(msg.fails)
		return
	}
	go func() {
		defer close(msg.fails)
		defer file.Close()
		if _, err := io.Copy(file, msg.contents); err != nil {
			msg.fails <- err
			return
		}
	}()
}

func (set *fileset1) onRemoveFile(msg *rremove) {
	defer close(msg.fails)
	if err := os.RemoveAll(msg.fpath); err != nil {
		msg.fails <- err
	}
}

func (set *fileset1) onLoadFile(msg *rload) {
	defer close(msg.results)
	file, err := os.OpenFile(msg.fpath, os.O_RDWR|os.O_SYNC, 0664)
	if err != nil {
		return
	}
	msg.results <- file
}

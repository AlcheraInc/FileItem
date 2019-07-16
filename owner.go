package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type owner2 struct {
	name, folder string
}

var (
	owner *owner2   // owner becomes a gateway to use the agent and mount
	agent itemAgent // representation of the items
	mount fsMount   // provides access control to the current file system
)

func init() {
	owner = new(owner2)
	owner.name = owner.GetName()
	wd, _ := os.Getwd()
	owner.folder = filepath.Join(wd, owner.name)
	os.Mkdir(owner.folder, 0775)

	agent.setup(7)
	go agent.start(owner)

	mount.setup(7)
	go mount.start(owner)
}

// GetDefaultItemOwner ...
//	Acquire pre-created item owner
//
// 	Todo
//		Current implementation requires it to be a singleton
//
func GetDefaultItemOwner() ItemOwner {
	if owner == nil {
		panic(errors.New("don't invoke this function in 'init()'"))
	}
	return owner
}

func (set *owner2) Close() error {
	agent.stop(nil)
	mount.stop(nil)
	return nil
}

func (set *owner2) GetName() (name string) {
	name, _ = os.Hostname() // use hostname by default
	return
}

func (set *owner2) GetPath() string {
	return set.folder
}

func (set *owner2) RemoveItems(items <-chan Item) {
	for item := range items {
		for err := range set.removeItem(item) {
			log.Println(err)
		}
	}
}

func (set *owner2) removeItem(item Item) <-chan error {
	var req rdelete
	req.Name, req.Type = item.GetName(), item.GetType()

	errs := make(chan error, 1)
	req.fails = errs
	agent.deletes <- req
	return errs
}

func (set *owner2) onDeleteItem(msg rdelete) {
	defer close(msg.fails)

	itempath := filepath.Join(set.folder, msg.GetType(), msg.GetName())
	if err := os.RemoveAll(itempath); err != nil {
		msg.fails <- err
	}
}

func (set *owner2) NewItem(iname, itype string) <-chan error {
	errs := make(chan error, 1)
	// prevent bad request with empty name and empty type
	if len(iname) == 0 || len(itype) == 0 {
		errs <- os.ErrInvalid
		close(errs)
		return errs
	}
	var req rcreate
	req.Name, req.Type = iname, itype
	req.fails = errs

	agent.creates <- req
	return errs
}

func (set *owner2) onCreateItem(msg rcreate) {
	defer close(msg.fails)
	// type's path
	//	we will allow multiple items are placed in the type.
	// 	just check if there is an existing file with the same name
	tpath := filepath.Join(set.folder, msg.GetType())
	if s, err := os.Stat(tpath); os.IsExist(err) {
		if s.IsDir() == false {
			msg.fails <- fmt.Errorf(
				"can't create type folder '%s'", msg.GetType())
		}
	}
	os.Mkdir(tpath, 0775)
	// item's path
	//	ban duplication
	ipath := filepath.Join(tpath, msg.GetName())
	if err := os.Mkdir(ipath, 0775); err != nil {
		if os.IsExist(err) {
			msg.fails <- fmt.Errorf("the item already exists")
			return
		}
		msg.fails <- err
	}
}

func (set *owner2) FindInType(itemtype string) <-chan string {
	names := make(chan string, 100)
	go set.enumerateDir(filepath.Join(set.folder, itemtype), names)
	return names
}

func (set *owner2) enumerateDir(fpath string, names chan<- string) {
	defer close(names)
	files, err := ioutil.ReadDir(fpath)
	if err != nil {
		log.Println(err)
		return
	}
	for _, f := range files {
		if f.IsDir() == false { // enumerateDir
			continue
		}
		names <- f.Name()
	}
}
func (set *owner2) enumerateFiles(fpath string, names chan<- string) {
	defer close(names)
	files, err := ioutil.ReadDir(fpath)
	if err != nil {
		log.Println(err)
		return
	}
	for _, f := range files {
		if f.IsDir() == true { // enumerateFiles
			continue
		}
		names <- f.Name()
	}
}

func (set *owner2) FindItem(iname, itype string) <-chan error {
	errs := make(chan error, 1)
	// prevent bad request with empty name and empty type
	if len(iname) == 0 || len(itype) == 0 {
		errs <- os.ErrInvalid
		close(errs)
		return errs
	}
	var req rsearch
	req.Name, req.Type = iname, itype
	req.fails = errs
	agent.searches <- req
	return errs
}

func (set *owner2) onSearchItem(msg rsearch) {
	defer close(msg.fails)

	ipath := filepath.Join(set.folder, msg.GetType(), msg.GetName())
	s, err := os.Stat(ipath)
	if os.IsNotExist(err) {
		msg.fails <- fmt.Errorf(
			"can't find the type/item '%s/%s'", msg.GetType(), msg.GetName())
		return
	}
	if s.IsDir() == false {
		msg.fails <- fmt.Errorf(
			"can't use the name for item '%s/%s'", msg.GetType(), msg.GetName())
		return
	}
}

func (set *owner2) UpdateOutline(item Item) <-chan error {
	errs := make(chan error, 1)
	var req rupdate
	req.item, req.fails = item, errs

	agent.updates <- req
	return errs
}

func (set *owner2) onUpdateItem(msg rupdate) {
	item := msg.item // alias to make code shorter

	var req faccess
	req.fpath = filepath.Join(set.folder, item.GetType(), item.GetName(), CacheFile)

	// create locally. and delegate the work to mount
	file, err := os.Create(req.fpath)
	if err != nil {
		req.fails <- err
		close(req.fails)
		return
	}
	file.Close()

	req.fails = msg.fails
	req.perm = 0664
	req.foper = func(file *os.File, err error) error {
		if err != nil {
			return err // redirect without processing
		}
		defer file.Close()
		enc := json.NewEncoder(file)
		return enc.Encode(item.GetOutline())
	}
	mount.accesses <- req
}

func (set *owner2) UseFile(item FileItem, fname string, operation FileOp) <-chan error {
	errs := make(chan error, 1)
	if len(fname) == 0 || operation == nil {
		errs <- os.ErrInvalid
		close(errs)
		return errs
	}
	var req faccess
	req.fpath = filepath.Join(item.GetPath(), fname)
	req.fails, req.foper = errs, operation
	req.perm = 775
	mount.accesses <- req
	return errs
}

package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"io"
	"path/filepath"
	"time"
)

//	simplest FileItem struct
type item1 struct {
	Name  string    `json:"name"`
	Type  string    `json:"resource_type"`
	Ctime time.Time `json:"created_time"`
	root  string    // `json:"root_path,omitempty"`
	owner *fileset1
}

func newitem1(iname, itype string) *item1 {
	r := new(item1)
	r.Name, r.Type = iname, itype
	return r
}

func (r *item1) toMap() map[string]interface{} {
	return map[string]interface{}{
		"name":          r.GetName(),
		"resource_type": r.GetType(),
		"created_time":  r.Ctime,
	}
}
func (r *item1) mergeTo(m map[string]interface{}) {
	m["name"] = r.GetName()
	m["resource_type"] = r.GetType()
	m["created_time"] = r.Ctime
}

func (r *item1) GetName() string {
	return r.Name
}
func (r *item1) GetType() string {
	return r.Type
}

func (r *item1) GetPath() string {
	return filepath.Join(r.root, r.GetType(), r.GetName())
}

func (r *item1) RemoveFile(fname, ftype string) <-chan error {
	errs := make(chan error)

	req := new(rremove)
	req.fpath = filepath.Join(r.GetPath(), fname)
	req.fails = errs
	r.owner.removes <- req

	return errs
}

func (r *item1) SaveFile(fname, ftype string, reader io.Reader) <-chan error {
	errs := make(chan error)

	req := new(rsave)
	req.contents = reader
	req.fpath = filepath.Join(r.GetPath(), fname)
	req.fails = errs
	r.owner.saves <- req

	return errs
}

func (r *item1) LoadFile(fname, ftype string) <-chan io.ReadWriteCloser {
	files := make(chan io.ReadWriteCloser)

	req := new(rload)
	req.results = files
	req.fpath = filepath.Join(r.GetPath(), fname)
	r.owner.loads <- req

	return files
}

func (r *item1) GetFiles() <-chan string {
	ch := make(chan string, 0)
	defer close(ch)
	return ch
}

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

func (r *item1) SaveFile(fname, ftype string, reader io.Reader) <-chan error {
	errs := make(chan error)

	req := new(rsave)
	req.contents = reader
	req.fpath = filepath.Join(r.GetPath(), fname)
	req.fails = errs
	r.owner.saves <- req
	// defer close(errs)
	// f, err := os.OpenFile(filepath.Join(r.GetPath(), fname), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0664)
	// if err != nil {
	// 	errs <- err
	// 	return
	// }
	// defer f.Close()
	// if _, err := io.Copy(f, reader); err != nil {
	// 	errs <- err
	// 	return
	// }

	return errs
}

func (r *item1) LoadFile(fname, ftype string) <-chan io.ReadWriteCloser {
	files := make(chan io.ReadWriteCloser)

	req := new(rload)
	req.results = files
	req.fpath = filepath.Join(r.GetPath(), fname)
	r.owner.loads <- req
	// defer close(files)
	// f, err := os.Open(filepath.Join(r.GetPath(), fname))
	// if err != nil {
	// 	return
	// }
	// files <- f

	return files
}

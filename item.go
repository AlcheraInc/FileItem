package fileitem

import "path/filepath"

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

// The simplest struct
type item2 struct {
	Name string `json:"name"`
	Type string `json:"resource_type"`
}

func (r *item2) GetName() string {
	return r.Name
}
func (r *item2) GetType() string {
	return r.Type
}

func (r *item2) GetOutline() interface{} {
	return r
}

type item3 struct {
	item2
	owner ItemOwner
}

func (item *item3) GetPath() string {
	return filepath.Join(item.owner.GetPath(), item.GetType(), item.GetName())
}

func (item *item3) GetFiles() <-chan string {
	fnames := make(chan string, 100)
	o := item.owner.(*owner2)
	go o.enumerateFiles(item.GetPath(), fnames)
	return fnames
}

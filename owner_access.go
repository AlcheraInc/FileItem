package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

// for now, use single goroutine access
var accesses map[string]interface{}

func init() {
	accesses = make(map[string]interface{})
}

func (set *owner2) inUse(fpath string) bool {
	_, exists := accesses[fpath]
	return exists
}
func (set *owner2) remember(fpath string) {
	accesses[fpath] = nil
}
func (set *owner2) forget(fpath string) {
	delete(accesses, fpath)
}

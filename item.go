package fileitem

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

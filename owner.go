package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"io"
)

// ItemFinder ...
//	Proxy to support item lookup
type ItemFinder interface {
	// FindNames ...
	//	Yield known items' name with given type
	FindNames(itemtype string) <-chan string

	// FindOne ...
	// 	Find one item with given name and type.
	//	Its detail will be loaded to receiver object
	FindOne(hint Item, receiver interface{}) <-chan error
}

// ItemCreator ...
//	Interface to create an FileItem or FileGroupItem
type ItemCreator interface {
	NewItem(iname, itype string, detail map[string]interface{}) FileGroupItem
}

// ItemOwner ...
//	Proxy to internal operation executor
type ItemOwner interface {
	io.Closer // explicit close is required
	ItemCreator
	ItemFinder

	GetPath() string

	// Remove ...
	//	Remove all owning resources
	Remove()

	// Delete ...
	//	Remove the given item
	Delete(item Item) <-chan error

	// Update ...
	//	Apply changes using map
	Update(item Item, changes map[string]interface{}) <-chan error
}
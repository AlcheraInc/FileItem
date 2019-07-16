package fileitem

import (
	"io"
	"os"
)

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

// Item ...
//	Ensure that the item has some properties
type Item interface {
	// GetName ...
	//	Return its own name for identification
	GetName() string

	// GetType ...
	//	Return its type for organization
	GetType() string

	// GetOutline ...
	//	Return some summary(outline) of the object
	GetOutline() interface{}
}

// CacheFile ...
//		Reserved JSON file to save the result from Item.GetOutline
const CacheFile = "info.json"

// FileItem ...
//	Item with the path
//	It can contain multiple resource in using its file system
type FileItem interface {
	Item

	// GetPath ...
	//	Current owning path of the item
	GetPath() string
	GetFiles() <-chan string
}

// ItemFinder ...
//	Proxy to support item lookup
type ItemFinder interface {
	// FindInType ...
	//	Yield known items' name with given type
	FindInType(itemtype string) <-chan string

	// FindItem ...
	//	Find with minimum hint.
	//	If the item doesn't exists, error will be delivered
	FindItem(iname, itype string) <-chan error
}

// ItemCreator ...
//	Interface to create an FileItem
type ItemCreator interface {
	// NewItem ...
	//	Create a space for the new item.
	//	If there is an existing item, error will be delivered
	NewItem(iname, itype string) <-chan error
}

// FileOp ...
//	Operation that is executed on a file
type FileOp func(*os.File, error) error

// ItemOwner ...
//	Proxy to internal operation executor
type ItemOwner interface {
	io.Closer // explicit close is required
	ItemCreator
	ItemFinder

	// GetName ...
	GetName() string

	// GetPath ...
	//	Current owning path of this owner
	GetPath() string

	// RemoveItems ...
	//	Remove the given items
	RemoveItems(items <-chan Item)

	// UpdateOutline ...
	//	Save the item in as a file.
	UpdateOutline(item Item) <-chan error

	// UseFile ...
	//  Perform file operation.
	//  By using this method user can prevent race condition on the file
	UseFile(item FileItem, fname string, operation FileOp) <-chan error
}

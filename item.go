package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"io"
)

// Item ...
//	Ensure that the item has some properties
type Item interface {
	// GetName ...
	//	Return its own name for identification
	GetName() string

	// GetType ...
	//	Return its type for organization
	GetType() string
}

// FileItem ...
//	Item with the path
type FileItem interface {
	Item

	// GetPath ...
	//	Current owning path of the item
	GetPath() string
}

// FileGroupItem ...
//	Resource with multiple files
type FileGroupItem interface {
	FileItem

	GetFiles() <-chan string

	// SaveFile ...
	//	Save as a file with given name.
	//	If the file needs to be listed in object detail, use non-zero length ftype
	SaveFile(fname, ftype string, r io.Reader) <-chan error

	// RemoveFile ...
	//	Delete a file it exists
	RemoveFile(fname, ftype string) <-chan error

	// LoadFile ...
	//	Load a file with given name and type
	LoadFile(fname, ftype string) <-chan io.ReadWriteCloser
}

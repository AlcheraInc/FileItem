package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"io"
)

// Item ...
//	Ensure item has type or organization and its own name for identification
type Item interface {
	GetName() string
	GetType() string
}

// FileItem ...
//	Item with the path
type FileItem interface {
	Item

	GetPath() string
}

// FileGroupItem ...
//  Resource with multiple files
type FileGroupItem interface {
	FileItem

	// SaveFile ...
	//	Save as a file with given name.
	//	If the file needs to be listed in object detail, use non-zero length ftype
	SaveFile(fname, ftype string, r io.Reader) <-chan error
	// LoadFile ...
	//	Load a file with given name and type
	LoadFile(fname, ftype string) <-chan io.ReadWriteCloser
}

package fileitem

import "testing"

func TestOwnerNotNull(t *testing.T) {
	owner := GetDefaultItemOwner()
	t.Log(owner.GetName())
	t.Log(owner.GetPath())
}

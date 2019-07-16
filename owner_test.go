package fileitem

import "testing"

func makeItemT(itype, iname string, t *testing.T) FileItem {
	i := new(item3)
	i.Type, i.Name = itype, iname
	i.owner = GetDefaultItemOwner()
	return i
}

func TestOwnerNotNull(t *testing.T) {
	owner := GetDefaultItemOwner()
	t.Log(owner.GetName())
	t.Log(owner.GetPath())
}

func TestOwnerDeleteNotExistingItem(t *testing.T) {
	owner := GetDefaultItemOwner()

	items := make(chan Item, 2)
	go owner.RemoveItems(items)
	defer close(items)
	items <- makeItemT("type-1", "name-unknown-1", t)
	items <- makeItemT("type-1", "name-unknown-2", t)
}

func TestOwnerCreateAndUpdateItem(t *testing.T) {
	owner := GetDefaultItemOwner()

	item := makeItemT("type-1", "name-1", t)
	for err := range owner.NewItem(item.GetName(), item.GetType()) {
		t.Fatal(err)
	}
	for err := range owner.UpdateOutline(item) {
		t.Fatal(err)
	}

	items := make(chan Item, 2)
	go owner.RemoveItems(items)
	defer close(items)
	items <- item
}

func TestOwnerProhibitDuplication(t *testing.T) {
	owner := GetDefaultItemOwner()

	item := makeItemT("type-1", "name-2", t)
	for err := range owner.NewItem(item.GetName(), item.GetType()) {
		t.Fatal(err)
	}
	for err := range owner.NewItem(item.GetName(), item.GetType()) {
		t.Log(err)
		return
	}
	t.FailNow()
}

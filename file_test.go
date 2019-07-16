package fileitem

import (
	"os"
	"testing"
)

func TestOwnerCreatesInfoJson(t *testing.T) {
	owner := GetDefaultItemOwner()

	item := makeItemT("type-2", "name-cache-user", t)

	for err := range owner.NewItem(item.GetName(), item.GetType()) {
		t.Fatal(err)
	}
	for err := range owner.UpdateOutline(item) {
		t.Fatal(err)
	}

	err := <-owner.UseFile(item, CacheFile, func(file *os.File, err error) error {
		if file != nil {
			file.Close()
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

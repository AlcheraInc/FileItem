package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"
)

func isFile(t *testing.T, p string) {
	stat, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if stat.IsDir() {
		t.FailNow()
	}
}

func isDirectory(t *testing.T, p string) {
	stat, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if stat.IsDir() == false {
		t.FailNow()
	}
}

func TestOwnerCreatesFolder(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	isDirectory(t, owner.GetPath())
}

func TestOwnerClosePreservesPath(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}

	isDirectory(t, owner.GetPath())
	owner.Remove()
}
func TestOwnerAllowNullDelete(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	for err := range owner.Delete(nil) {
		t.Fatal(err)
	}
}

func TestOwnerSearchWithType(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item1 := <-owner.NewItem("item1", "test_item", nil)
	item2 := <-owner.NewItem("item2", "test_item", nil)
	item3 := <-owner.NewItem("item3", "test_item2", nil)
	item4 := <-owner.NewItem("item4", "test_item2", nil)

	count := 0
	for name := range owner.FindNames("test_item") {
		if name == item1.GetName() || name == item2.GetName() {
			count = count + 1
		}
	}
	if count < 2 {
		t.FailNow()
	}
	count = 0
	for name := range owner.FindNames("test_item2") {
		if name == item3.GetName() || name == item4.GetName() {
			count = count + 1
		}
	}
	if count < 2 {
		t.FailNow()
	}
}
func TestOwnerSearchWithUnknownType(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	// given: empty owner
	// when: search with unknwon type name
	for range owner.FindNames("unknown_type_name") {
		// then: nothing returns
		t.FailNow()
	}
}

func TestOwnerFindExistingItem(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item1 := <-owner.NewItem("item5", "test_item", nil)
	if item1 == nil {
		t.FailNow()
	}
	t.Log(item1.GetName(), item1.GetType())
	time.Sleep(1 * time.Second)

	for range owner.Find(item1.GetName(), item1.GetType()) {
		return
	}
	t.Fatal("Find after NewItem failed")
}

func TestOwnerFindUnknownItem(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item := owner.Find("item19", "test_item")
	if item == nil {
		t.Fatal("Find after NewItem failed")
	}
}

func TestOwnerFailsToFindDeletedItem(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item1 := <-owner.NewItem("item7", "test_item", nil)
	if item1 == nil {
		t.FailNow()
	}

	for err := range owner.Delete(item1) {
		t.Fatal(err)
	}

	for range owner.Find(item1.GetName(), item1.GetType()) {
		t.FailNow()
	}
}

type described struct {
	Name string `json:"name"`
	Desc string `json:"description"`
}

func TestOwnerFindItemAndUpdate(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item1 := <-owner.NewItem("item8", "test_item", map[string]interface{}{
		"description": "hell world",
	})
	if item1 == nil {
		t.FailNow()
	}

	data := new(described)

	item1 = <-owner.Find(item1.GetName(), item1.GetType())

	for file := range item1.LoadFile(CacheFile, "") {
		err := json.NewDecoder(file).Decode(&data)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	if data.Desc != "hell world" {
		t.FailNow()
	}
	for err := range owner.Update(item1, map[string]interface{}{
		"description": "world hell",
	}) {
		t.Fatal(err)
	}

	var data2 map[string]interface{}
	file2 := <-item1.LoadFile(CacheFile, "")
	defer file2.Close()
	if err := json.NewDecoder(file2).Decode(&data2); err != nil {
		t.Fatal(err)
	}
	if data.Desc != "hell world" {
		t.FailNow()
	}
	if data2["resource_type"] == nil {
		t.FailNow()
	}
	if data2["description"].(string) != "world hell" {
		t.FailNow()
	}
}

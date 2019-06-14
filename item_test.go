package fileitem

//
//	Authors
//		github.com/luncliff	(dh.park@alcherainc.com)
//

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

type tagged struct {
	Name  string    `json:"name"`
	Type  string    `json:"resource_type"`
	Ctime time.Time `json:"created_time"`
	Tag   []string  `json:"tags"`
}

func TestItemMarshalwithJSON(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item1 := <-owner.NewItem("item9", "test_item", map[string]interface{}{
		"tags": []string{"tag1", "tag2"},
	})
	if item1 == nil {
		t.FailNow()
	}

	item := new(tagged)
	for file := range item1.LoadFile(CacheFile, "") {
		err := json.NewDecoder(file).Decode(&item)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	if item.Name != item1.GetName() {
		t.Fatal("name mismatch", item.Name, item1.GetName())
	}
	if item.Tag == nil || len(item.Tag) == 0 {
		t.FailNow()
	}
	b, _ := json.MarshalIndent(item, "    ", "    ")
	t.Log(string(b))
}

func TestCreateNewItemAllNull(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item := owner.NewItem("item10", "test_item", nil)
	if item == nil {
		t.FailNow()
	}
}

func TestCreateNewItemMakesFolder(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item := <-owner.NewItem("item11", "test_item", nil)
	if item == nil {
		t.FailNow()
	}
	if item.GetName() != "item11" || item.GetType() != "test_item" {
		t.Fatal("name or type mismatch")
	}
	isDirectory(t, item.GetPath())
}

func prepareSysInfoT(t *testing.T) []byte {
	args := []string{}
	switch runtime.GOOS {
	case "windows":
		args = []string{"systeminfo"}
	default:
		args = []string{"uname", "-a"}
	}
	sub := exec.Command(args[0], args[1:]...)
	rc, _ := sub.StdoutPipe()
	if err := sub.Start(); err != nil {
		t.Fatal(err)
	}
	defer sub.Wait()
	defer rc.Close()
	blob, err := ioutil.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	return blob
}

func TestItemAllowMultipleSave(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item := <-owner.NewItem("item12", "test_item", nil)
	if item == nil {
		t.FailNow()
	}

	blob := prepareSysInfoT(t)
	for err := range item.SaveFile("hello.txt", "", bytes.NewBuffer(blob)) {
		t.Fatal(err)
	}
	for err := range item.SaveFile("hello.txt", "", bytes.NewBuffer(blob)) {
		t.Fatal(err)
	}
	exists := false
	for fname := range item.GetFiles() {
		t.Log(fname)
		if fname == "hello.txt" {
			exists = true
		}
	}
	if exists == false {
		t.FailNow()
	}
}

func TestItemLoadUnknownFile(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item := <-owner.NewItem("item13", "test_item", nil)
	if item == nil {
		t.FailNow()
	}

	rc := <-item.LoadFile("unknwon-file.txt", "")
	if rc != nil {
		defer rc.Close()
		t.FailNow()
	}
}

func TestItemLoadAfterSave(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item := <-owner.NewItem("item14", "test_item", nil)
	if item == nil {
		t.FailNow()
	}

	blob := prepareSysInfoT(t)
	for err := range item.SaveFile("hello.txt", "", bytes.NewBuffer(blob)) {
		t.Fatal(err)
	}

	rc := <-item.LoadFile("hello.txt", "")
	if rc == nil {
		t.FailNow()
	}
	defer rc.Close()
	blob2, _ := ioutil.ReadAll(rc)

	if bytes.Equal(blob, blob2) == false {
		t.FailNow()
	}
}

func TestItemDeleteUnknownFile(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item := <-owner.NewItem("item15", "test_item", nil)
	if item == nil {
		t.FailNow()
	}
	for err := range item.RemoveFile("hello.txt", "") {
		t.Fatal(err)
	}
}

func TestItemLoadAfterDelete(t *testing.T) {
	owner, err := NewOwner1("pkg")
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()

	item := <-owner.NewItem("item16", "test_item", nil)
	if item == nil {
		t.FailNow()
	}

	blob := prepareSysInfoT(t)
	for err := range item.SaveFile("hello.txt", "", bytes.NewBuffer(blob)) {
		t.Fatal(err)
	}

	for err := range item.RemoveFile("hello.txt", "") {
		t.Fatal(err)
	}

	rc := <-item.LoadFile("hello.txt", "")
	if rc != nil {
		t.FailNow()
	}
}

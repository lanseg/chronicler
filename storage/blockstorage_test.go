package storage

import (
	"reflect"
	"testing"
)

type someType struct {
	String string
	Array  []*someType
}

func TestBlockStorage(t *testing.T) {
	s, err := NewLocalStorage(t.TempDir())
	if err != nil {
		t.Errorf("Cannot create temporary storage: %s", err)
	}

	bs := &BlockStorage{Storage: s}
	t.Run("read write object", func(t *testing.T) {
		want := &someType{
			String: "Root",
			Array:  []*someType{{String: "Child"}},
		}
		if _, err := bs.PutObject(&PutRequest{Url: "obj-file"}, want); err != nil {
			t.Errorf("Error writing object to storage: %s", err)
		}
		got := &someType{}
		if err = bs.GetObject(&GetRequest{Url: "obj-file"}, got); err != nil {
			t.Errorf("Error reading object from storage: %s", err)
		}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("Expected GetJson result to be %v, but got %v", want, got)
		}
	})

	t.Run("read write bytes", func(t *testing.T) {
		toWrite := []byte{1, 2, 3, 4, 5}
		if _, err := bs.PutBytes(&PutRequest{Url: "byte-file"}, toWrite); err != nil {
			t.Errorf("Error writing json to storage: %s", err)
		}
		var toRead []byte
		if toRead, err = bs.GetBytes(&GetRequest{Url: "byte-file"}); err != nil {
			t.Errorf("Error reading json from storage: %s", err)
		}
		if !reflect.DeepEqual(toRead, toWrite) {
			t.Errorf("Expected GetBytes result to be %s, but got %s", toRead, toWrite)
		}
	})
}

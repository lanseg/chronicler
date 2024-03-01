package storage

import (
	"fmt"
	"io"
	"reflect"
	"testing"
)

func staticId(id string) IdSource {
	src := 0
	return func() string {
		src++
		return fmt.Sprintf("%s_%d", id, src)
	}
}

type sampleFile struct {
	originalFileName string
	content          []byte
}

func TestOverlay(t *testing.T) {

	for _, tc := range []struct {
		name          string
		originalFiles []*sampleFile
		expectedFiles []*sampleFile
	}{
		{
			name: "Single file name",
			originalFiles: []*sampleFile{{
				originalFileName: "Whatever",
				content:          []byte{1, 2, 3},
			}},
		},
		{
			name: "File name with spaces",
			originalFiles: []*sampleFile{{
				originalFileName: "A spaces file name",
				content:          []byte{1, 2, 3, 4, 5},
			}},
		},
		{
			name: "File name with unicode",
			originalFiles: []*sampleFile{{
				originalFileName: "Немного ユニコード",
				content:          []byte{1, 2, 3, 4, 5},
			}},
		},
		{
			name: "File name with path",
			originalFiles: []*sampleFile{{
				originalFileName: "some/file/path",
				content:          []byte{1, 2, 3, 4, 5},
			}},
		},
		{
			name: "Empty file",
			originalFiles: []*sampleFile{{
				originalFileName: "some empty file",
				content:          []byte{},
			}},
		},
		{
			name: "Two files different names",
			originalFiles: []*sampleFile{
				{"File 1", []byte{1, 2, 3}},
				{"File 2", []byte{4, 5, 6}},
			},
		},
		{
			name: "Two files same name",
			originalFiles: []*sampleFile{
				{"File 1", []byte{1, 2, 3}},
				{"File 1", []byte{4, 5, 6}},
			},
			expectedFiles: []*sampleFile{
				{"File 1", []byte{4, 5, 6}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := NewOverlay(t.TempDir(), staticId("SomeId"))
			for _, f := range tc.originalFiles {
				o.Write(f.originalFileName, f.content)
			}

			if tc.expectedFiles == nil {
				tc.expectedFiles = tc.originalFiles
			}
			result := []*sampleFile{}
			for _, f := range tc.expectedFiles {
				reader, err := o.Read(f.originalFileName).Get()
				content := []byte{}
				if err == nil {
					content, err = io.ReadAll(reader)
				}
				if err == nil {
					result = append(result, &sampleFile{f.originalFileName, content})
				}
			}

			if !reflect.DeepEqual(result, tc.expectedFiles) {
				t.Errorf("Expected result to be %s, but got %s", tc.expectedFiles, result)
			}
		})
	}
}
